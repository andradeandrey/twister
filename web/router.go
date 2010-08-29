// Copyright 2010 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package web

import (
	"bytes"
	"container/vector"
	"http"
	"regexp"
	"utf8"
	"flag"
	"strings"
)

// Router dispatches HTTP requests to a handler using the path component of the
// request URL and the request method.
//
// A router maintains a list of routes. A route consists of a request path
// pattern and a collection of (method, handler) pairs.
//
// A pattern is a string with embedded parameters. A parameter has the syntax:
//
// '<' name (':' regexp)? '>'
//
// If the regexp is not specified, then the regexp is set to to [^/X]+ where
// "X" is the character following the closing '>' or nothing if the closing
// '>' is at the end of the pattern.
//
// The pattern must begin with the character '/'.
//
// A router dispatches requests by matching the path component of the request
// URL against the route patterns in the order that the routes were registered.
// If a matching route is found, then the router searches the route for a
// handler using the request method, "GET" if the request method is "HEAD" and
// "*". If a handler is not found, the router responds with HTTP status 405. If
// a route is not found, then the router responds with HTTP status 404.
//
// The handler can access the path parameters in the request Form.
//
// If a pattern ends with '/', then the router redirects the URL without the
// trailing slash to the URL with the trailing slash.
//
type Router struct {
	errorHandler func(statusCode int, conn *http.Conn, req *http.Request)
	routes       vector.Vector
}


type route struct {
	addSlash bool
	regexp   *regexp.Regexp
	names    []string
	handlers map[string]http.Handler
}

var parameterRegexp = regexp.MustCompile("<([A-Za-z0-9]+)(:[^>]*)?>")

// compilePattern compiles the pattern to a regexp and array of paramter names.
func compilePattern(pattern string, addSlash bool) (*regexp.Regexp, []string) {
	var buf bytes.Buffer
	names := make([]string, 8)
	i := 0
	buf.WriteString("^")
	for {
		a := parameterRegexp.FindStringSubmatchIndex(pattern)
		if len(a) == 0 {
			buf.WriteString(regexp.QuoteMeta(pattern))
			break
		} else {
			buf.WriteString(regexp.QuoteMeta(pattern[0:a[0]]))
			names[i] = pattern[a[2]:a[3]]
			i += 1
			if a[4] >= 0 {
				buf.WriteString("(")
				buf.WriteString(pattern[a[4]+1 : a[5]])
				buf.WriteString(")")
			} else {
				buf.WriteString("([^")
				if a[1] < len(pattern) {
					rune, _ := utf8.DecodeRuneInString(pattern[a[1]:])
					if rune != '/' {
						buf.WriteRune(rune)
					}
				}
				buf.WriteString("/]+)")
			}
			pattern = pattern[a[1]:]
		}
	}
	if addSlash {
		buf.WriteString("?")
	}
	buf.WriteString("$")
	return regexp.MustCompile(buf.String()), names[0:i]
}

// Register the route with the given pattern and handlers. The structure of the
// handlers argument is:
//
// [method handler]+
//
// where method is a string and handler is an http.Handler or a
// func(*http.Conn, *http.Request). Use "*" to match all methods.
func (router *Router) Register(pattern string, handlers ...interface{}) *Router {
	if pattern == "" || pattern[0] != '/' {
		panic("twister: Invalid route pattern " + pattern)
	}
	if len(handlers)%2 != 0 || len(handlers) == 0 {
		panic("twister: Invalid handlers for pattern " + pattern +
			". Structure of handlers is [method handler]+.")
	}
	r := route{}
	r.addSlash = pattern[len(pattern)-1] == '/'
	r.regexp, r.names = compilePattern(pattern, r.addSlash)
	r.handlers = make(map[string]http.Handler)
	for i := 0; i < len(handlers); i += 2 {
		method, ok := handlers[i].(string)
		if !ok {
			panic("twister: Bad method for pattern " + pattern)
		}
		switch handler := handlers[i+1].(type) {
		case http.Handler:
			r.handlers[method] = handler
		case func(*http.Conn, *http.Request):
			r.handlers[method] = http.HandlerFunc(handler)
		default:
			panic("twister: Bad handler for pattern " + pattern + " and method " + method)
		}
	}
	router.routes.Push(&r)
    return router
}

type routerError struct {
	router     *Router
	statusCode int
    message     string
}

func (re *routerError) ServeHTTP(conn *http.Conn, req *http.Request) {
    if re.router.errorHandler != nil  {
        re.router.errorHandler(re.statusCode, conn, req)
    } else {
        http.Error(conn, re.message, re.statusCode)
    }
}

// addSlash redirects to the request URL with a trailing slash.
func addSlash(conn *http.Conn, req *http.Request) {
	path := req.URL.Path + "/"
	if len(req.URL.RawQuery) > 0 {
		path = path + "?" + req.URL.RawQuery
	}
	http.Redirect(conn, path, http.StatusMovedPermanently)
}

// Given the path componennt of the request URL and the request method, find
// the handler and path parameters.
func (router *Router) find(path string, method string) (http.Handler, []string, []string) {
	for i := 0; i < router.routes.Len(); i++ {
		r := router.routes.At(i).(*route)
		values := r.regexp.FindStringSubmatch(path)
		if len(values) == 0 {
			continue
		}
		if r.addSlash && path[len(path)-1] != '/' {
			return http.HandlerFunc(addSlash), nil, nil
		}
		values = values[1:]
		for j := 0; j < len(values); j++ {
			if value, e := http.URLUnescape(values[j]); e != nil {
				return &routerError{router, 400, "Bad request."}, nil, nil
			} else {
				values[j] = value
			}
		}
		if handler := r.handlers[method]; handler != nil {
			return handler, r.names, values
		}
		if method == "HEAD" {
			if handler := r.handlers["GET"]; handler != nil {
				return handler, r.names, values
			}
		}
		if handler := r.handlers["*"]; handler != nil {
			return handler, r.names, values
		}
		return &routerError{router, 405, "Method not supported."}, nil, nil
	}
	return &routerError{router, 404, "Not found."}, nil, nil
}

// ServeHTTP dispatches the request to a registered handler.
func (router *Router) ServeHTTP(conn *http.Conn, req *http.Request) {
	handler, names, values := router.find(req.URL.Path, req.Method)
	req.ParseForm()
	for i := 0; i < len(names); i++ {
		req.Form[names[i]] = []string{values[i]}
	}
	handler.ServeHTTP(conn, req)
}

// NewRouter allocates and initializes a new Router. If the router encounters
// an error when serving a request, then the errorHandler function is called to
// generate the response.
func NewRouter(errorHandler func(statusCode int, conn *http.Conn, req *http.Request)) *Router {
	return &Router{errorHandler: errorHandler}
}

// HostRouter dispatches HTTP requests to a handler using the host header.
//
// To enable debugging on localhost, the router overrides the request host with
// the value of the hostOverride flag if set.
//
// If a registered handler is not found, then the router dispatches to a
// default handler.
type HostRouter struct {
	defaultHandler http.Handler
	handlers       map[string]http.Handler
}

// NewHostRouter allocates and initializes a new HostRouter.
func NewHostRouter(defaultHandler http.Handler) *HostRouter {
    if defaultHandler == nil {
        defaultHandler = http.NotFoundHandler()
    }
	return &HostRouter{defaultHandler: defaultHandler, handlers: make(map[string]http.Handler)}
}

// Register a handler for the given host.
func (router *HostRouter) Register(host string, handler http.Handler) *HostRouter {
	router.handlers[strings.ToLower(host)] = handler
    return router
}

var hostOverride = flag.String("hostOverride", "", "Override request host in HostRouter")

// ServeHTTP dispatches the request to a registered handler.
func (router *HostRouter) ServeHTTP(conn *http.Conn, req *http.Request) {
	var host string
	if len(*hostOverride) == 0 {
		host = strings.ToLower(req.Host)
	} else {
		host = *hostOverride
	}
	if handler, found := router.handlers[host]; found {
		handler.ServeHTTP(conn, req)
	} else {
		router.defaultHandler.ServeHTTP(conn, req)
	}
}
