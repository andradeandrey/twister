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
	"container/vector"
	"http"
	"fmt"
	"io"
	"os"
	"bufio"
)

// StringsMap maps strings to slices of strings.
type StringsMap map[string][]string

// Get returns first value for given key.
func (m StringsMap) Get(key string) (value string, found bool) {
	values, found := m[key]
	if found && len(values) > 0 {
		return value, true
	}
	return "", false
}

// GetDef returns first value for given key, or def if key is not found.
func (m StringsMap) GetDef(key string, def string) (value string) {
	values, found := m[key]
	if found && len(values) > 0 {
		return value
	}
	return def
}

// Append value to slice for given key.
func (m StringsMap) Append(key string, value string) {
	v := vector.StringVector(m[key])
	v.Push(value)
	m[key] = v
}

// Set value for given key, discarding previous values if any.
func (m StringsMap) Set(key string, value string) {
	m[key] = []string{value}
}

type RequestBody interface {
	io.Reader
}

type ResponseBody interface {
	io.Writer
	// Flush writes any buffered data to the network.
	Flush() os.Error
}

// Connection represents the server side of an HTTP connection.
type Connection interface {
	// Respond commits the status and headers to the network and returns
	// a writer for the response body.
	Respond(status int, header StringsMap) ResponseBody

	// Hijack lets the caller take over the connection from the HTTP server.
	// The caller is responsible for closing the connection.
	Hijack() (rwc io.ReadWriteCloser, buf *bufio.ReadWriter, err os.Error)
}

// Request represents an HTTP request.
type Request struct {
	Connection    Connection        // The connection.
	Method        string            // Uppercase request method. GET, POST, etc.
	URL           *http.URL         // Parsed URL.
	ProtocolMajor int               // HTTP major version.
	ProtocolMinor int               // HTTP minor version.
	Param         StringsMap        // Form, QS and other.
	Cookie        map[string]string // Cookies.
	Host          string

	// ErrorHandler responds to the request with the given status code.
	// Applications set their error handler in middleware. 
	ErrorHandler func(req *Request, status int, message string)

	// Header maps canonical header names to slices of header values.
	Header StringsMap

	// ContentLength is the length of the request body. If the value is -1,
	// then the length of the request body is not known.
	ContentLength int

	Body RequestBody
}

// Handler is the interface for web handlers.
type Handler interface {
	ServeWeb(req *Request)
}

// HandlerFunc is a type adapter to allow the use of ordinary functions as web
// handlers.
type HandlerFunc func(*Request)

// ServeWeb calls f(req).
func (f HandlerFunc) ServeWeb(req *Request) { f(req) }

// NewRequest allocates and initializes an empty request.
func NewRequest() *Request {
	return &Request{
		ContentLength: -1,
		ErrorHandler:  defaultErrorHandler,
		Param:         make(map[string][]string),
		Cookie:        make(map[string]string),
		Header:        make(map[string][]string),
	}
}

// Respond is a convenience function that adds (key, value) pairs in kvs to a
// StringsMap and calls through to the connection's Respond method.
func (req *Request) Respond(status int, kvs ...string) ResponseBody {
	if len(kvs)%2 == 1 {
		panic("twister: respond requires even number of kvs args")
	}
	header := StringsMap(make(map[string][]string))
	for i := 0; i < len(kvs); i += 2 {
		header.Append(kvs[i], kvs[i+1])
	}
	return req.Connection.Respond(status, header)
}

func defaultErrorHandler(req *Request, status int, message string) {
	w := req.Respond(status, HeaderContentType, "text/plain; charset=utf-8")
	if w != nil {
		fmt.Fprintln(w, message)
	}
}

// Error responds to the request with an error. 
func (req *Request) Error(status int, message string) {
	req.ErrorHandler(req, status, message)
}

// Redirect responds to the request with a redirect the specified URL.
func (req *Request) Redirect(url string, perm bool) {
	status := StatusFound
	if perm {
		status = StatusMovedPermanently
	}
	// TODO fix up url
	req.Respond(status, HeaderLocation, url)
}

type redirectHandler struct {
	url       string
	permanent bool
}

func (rh *redirectHandler) ServeWeb(req *Request) {
	req.Redirect(rh.url, rh.permanent)
}

// RedirectHandler returns a request handler that redirects to the given URL. 
func RedirectHandler(url string, permanent bool) Handler {
	return &redirectHandler{url, permanent}
}

var notFoundHandler = HandlerFunc(func(req *Request) { req.Error(StatusNotFound, "Not Found") })

// NotFoundHandler returns a request handler that responds with 404 not found.
func NotFoundHandler() Handler {
	return notFoundHandler
}

// HeaderName returns the canonical format of the header name s. 
func HeaderName(name string) string {
	p := []byte(name)
	return HeaderNameBytes(p)
}

// HeaderNameBytes returns the canonical format for the header name specified
// by the bytes in p. This function modifies the contents p.
func HeaderNameBytes(p []byte) string {
	upper := true
	for i, c := range p {
		if upper {
			if 'a' <= c && c <= 'z' {
				p[i] = c + 'A' - 'a'
			}
		} else {
			if 'A' <= c && c <= 'Z' {
				p[i] = c + 'a' - 'A'
			}
		}
		upper = c == '-'
	}
	return string(p)
}
