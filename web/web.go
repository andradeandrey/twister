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
	"http"
    "container/vector"
    "io"
    "strings"
)

var (
	isText [256]bool
	isToken  [256]bool
	isSpace  [256]bool
)

func init() {
	// From RFC 2616:
	// 
	// OCTET      = <any 8-bit sequence of data>
	// CHAR       = <any US-ASCII character (octets 0 - 127)>
	// CTL        = <any US-ASCII control character (octets 0 - 31) and DEL (127)>
	// CR         = <US-ASCII CR, carriage return (13)>
	// LF         = <US-ASCII LF, linefeed (10)>
	// SP         = <US-ASCII SP, space (32)>
	// HT         = <US-ASCII HT, horizontal-tab (9)>
	// <">        = <US-ASCII double-quote mark (34)>
	// CRLF       = CR LF
	// LWS        = [CRLF] 1*( SP | HT )
	// TEXT       = <any OCTET except CTLs, but including LWS>
	// separators = "(" | ")" | "<" | ">" | "@" | "," | ";" | ":" | "\" | <"> 
	//              | "/" | "[" | "]" | "?" | "=" | "{" | "}" | SP | HT
	// token      = 1*<any CHAR except CTLs or separators>
	// qdtext     = <any TEXT except <">>

	for c := 0; c < 256; c++ {
		isCtl := (0 <= c && c <= 31) || c == 127
		isChar := 0 <= c && c <= 127
		isSpace[c] = strings.IndexRune(" \t\r\n", c) >= 0
		isSeparator := strings.IndexRune(" \t\"(),/:;<=>?@[]\\{}", c) >= 0
		isText[c] = isSpace[c] || !isCtl
		isToken[c] = isChar && !isCtl && !isSeparator
	}
}

// IsTokenByte returns true if c is a token characeter as defined by RFC 2616
func IsTokenByte(c byte) bool {
    return isToken[c]
}

// IsSpaceByte returns true if c is a space characeter as defined by RFC 2616
func IsSpaceByte(c byte) bool {
    return isSpace[c]
}

// Headers is a map of canonical header keys to an array of strings. Use
type Headers map[string][]string

// Add value to list of headers for key. The key must be in canonical format.
func (h Headers) Add(key string, value string) {
	v := vector.StringVector(h[key])
	v.Push(value)
	h[key] = v
}

// Set header to value, replacing previous values if any. The key must be in
// canonical format.
func (h Headers) Set(key string, value string) {
	h[key] = []string{value}
}

// Get first header for key. The key must be in canonical format.
func (h Headers) Get(key string) (value string, ok bool) {
	values, ok := h[key]
	if ok {
		value = values[0]
		return value, true
	}
	return "", false
}

// CanonicalizeHeaderKeyBytes updates the argument to canonical key format were the
// words separated by '-' are capitalized.
func CanonicalizeHeaderKeyBytes(p []byte) {
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
}

// CanonicalHeaderKey returns the canonical format of the header key s.
func CanonicalHeaderKey(s string) string {
	// Other frameworks memoize this function, but that's not a good idea
	// because that allows an attacker to consume and arbitrary amount of
	// memory on the server.
	p := []byte(s)
	CanonicalizeHeaderKeyBytes(p)
	return string(p)
}

// Request represents a parsed HTTP request. 
type Request struct {
	// Request method, uppercase
	Method string
	// Parsed URL
	URL *http.URL
	// HTTP major version
	ProtocolMajor int
	// HTTP minor version
	ProtocolMinor int

	// Headers 
	Headers Headers

	// The message body.
	Body io.ReadCloser
}

