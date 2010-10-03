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
	"fmt"
	"http"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"net"
)

var (
	// Object not in valid state for call.
	ErrInvalidState = os.NewError("invalid state")
	ErrBadFormat    = os.NewError("bad format")
	errParsed       = os.NewError("item parsed")
)

// StringsMap maps strings to slices of strings.
type StringsMap map[string][]string

// NewStringsMap returns a map initialized with the given key-value pairs.
func NewStringsMap(kvs ...string) StringsMap {
	if len(kvs)%2 == 1 {
		panic("twister: even number args required for NewStringsMap")
	}
	m := make(StringsMap)
	for i := 0; i < len(kvs); i += 2 {
		m.Append(kvs[i], kvs[i+1])
	}
	return m
}

// Get returns first value for given key.
func (m StringsMap) Get(key string) (string, bool) {
	values, found := m[key]
	if !found || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// GetDef returns first value for given key, or def if key is not found.
func (m StringsMap) GetDef(key string, def string) string {
	values, found := m[key]
	if !found || len(values) == 0 {
		return def
	}
	return values[0]
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

// Responder represents the the response.
type Responder interface {
	// Respond commits the status and headers to the network and returns
	// a writer for the response body.
	Respond(status int, header StringsMap) ResponseBody

	// Hijack lets the caller take over the connection from the HTTP server.
	// The caller is responsible for closing the connection. 
	Hijack() (conn net.Conn, buffered []byte, err os.Error)
}

// Request represents an HTTP request.
type Request struct {
	Responder       Responder  // The response.
	Method          string     // Uppercase request method. GET, POST, etc.
	URL             *http.URL  // Parsed URL.
	ProtocolVersion int        // Major version * 1000 + minor version
	Param           StringsMap // Form, QS and other.
	Cookie          StringsMap // Cookies.
	Host            string     // Requested host.
	ContentType     string     // Lowercase content type, not including params

	// ErrorHandler responds to the request with the given status code.
	// Applications set their error handler in middleware. 
	ErrorHandler func(req *Request, status int, message string)

	// Header maps canonical header names to slices of header values.
	Header StringsMap

	// ContentLength is the length of the request body or -1 if the content
	// length is not known.
	ContentLength int

	Body RequestBody // The request body

	formParseErr os.Error
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

// NewRequest allocates and initializes a request.
func NewRequest(method string, url string, protocolVersion int, header StringsMap) (req *Request, err os.Error) {
	req = &Request{
		Method:          strings.ToUpper(method),
		ProtocolVersion: protocolVersion,
		ErrorHandler:    defaultErrorHandler,
		Param:           make(StringsMap),
		Header:          header,
		Cookie:          parseCookieValues(header[HeaderCookie]),
	}

	req.URL, err = http.ParseURL(url)
	if err != nil {
		return nil, err
	}

	err = parseUrlEncodedFormBytes([]byte(req.URL.RawQuery), req.Param)
	if err != nil {
		return nil, err
	}

	// The url overrides header per section 5.2 of rfc 2616
	req.Host = req.URL.Host
	if req.Host == "" {
		req.Host = req.Header.GetDef(HeaderHost, "")
	}

	if s, found := req.Header.Get(HeaderContentLength); found {
		var err os.Error
		req.ContentLength, err = strconv.Atoi(s)
		if err != nil {
			return nil, os.NewError("bad content length")
		}
	} else if method != "HEAD" && method != "GET" {
		req.ContentLength = -1
	}

	if s, found := req.Header.Get(HeaderContentType); found {
		i := 0
		for i < len(s) && (IsTokenByte(s[i]) || s[i] == '/') {
			i++
		}
		req.ContentType = strings.ToLower(s[0:i])
	}

	return req, nil
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
	return req.Responder.Respond(status, header)
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

	// Make relative path absolute
	u, err := http.ParseURL(url)
	if err != nil && u.Scheme == "" && url[0] != '/' {
		d, _ := path.Split(req.URL.Path)
		url = d + url
	}

	req.Respond(status, HeaderLocation, url)
}

// BodyBytes returns the request body a slice of bytees.
func (req *Request) BodyBytes() ([]byte, os.Error) {
	var p []byte
	if req.ContentLength > 0 {
		p = make([]byte, req.ContentLength)
		if _, err := req.Body.Read(p); err != nil {
			return nil, err
		}
	} else {
		var err os.Error
		if p, err = ioutil.ReadAll(req.Body); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// ParseForm parses url-encoded form bodies. ParseForm is idempotent.
func (req *Request) ParseForm() os.Error {
	if req.formParseErr == errParsed {
		return nil
	} else if req.formParseErr != nil {
		return req.formParseErr
	}
	req.formParseErr = errParsed
	if req.ContentType != "application/x-www-form-urlencoded" ||
		req.ContentLength == 0 ||
		(req.Method != "POST" && req.Method != "PUT") {
		return nil
	}
	p, err := req.BodyBytes()
	if err != nil {
		req.formParseErr = err
		return err
	}
	if err := parseUrlEncodedFormBytes(p, req.Param); err != nil {
		req.formParseErr = err
		return err
	}
	return nil
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

type Cookie struct {
	Name     string
	Value    string
	MaxAge   int
	Path     string
	Domain   string
	HttpOnly bool
	Secure   bool
}

func (c *Cookie) String() string {
	var b bytes.Buffer
	b.WriteString(c.Name)
	b.WriteRune('=')
	b.WriteString(c.Value)
	if c.MaxAge < 0 {
		// A date in the past will delete the cookie.
		b.WriteString("; Expires=Mon, 02 Jan 2006 15:04:05 GMT")
	}
	if c.MaxAge > 0 {
		// Write expires attribute because some browsers do not support max-age.
		b.WriteString("; Expires=")
		b.WriteString(time.SecondsToUTC(time.Seconds() + int64(c.MaxAge)).Format(TimeLayout))
	}
	if c.Path != "" {
		b.WriteString("; Path=")
		b.WriteString(c.Path)
	}
	if c.Domain != "" {
		b.WriteString("; Domain=")
		b.WriteString(c.Domain)
	}
	if c.Secure {
		b.WriteString("; Secure")
	}
	if c.HttpOnly {
		b.WriteString("; HttpOnly")
	}
	return b.String()
}
