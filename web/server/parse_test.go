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

package server

import (
	"testing"
	"bufio"
	"bytes"
	"reflect"
	"http"
	"github.com/garyburd/twister/web"
)

func compare(t *testing.T, expected *web.Request, s string) {
	b := bufio.NewReader(bytes.NewBufferString(s))
	actual, err := parseRequest(b)
	if expected == nil {
		if err == nil {
			t.Errorf("parse error expected")
			return
		}
	} else if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	} else if !reflect.DeepEqual(expected, actual) {
		t.Errorf("request not equal,\nexpected %q\nactual   %q", expected, actual)
	}
}

func mustParseURL(s string) *http.URL {
	url, err := http.ParseURL(s)
	if err != nil {
		panic(err)
	}
	return url
}

func TestEmpty(t *testing.T) {
	compare(t, nil, ` `)
}

func TestNoBody(t *testing.T) {
	compare(t, nil, "GET /foo HTTP/1.1\r\n")
}

func TestSimple(t *testing.T) {
	compare(t,
		&web.Request{
			ProtocolMajor: 1,
			ProtocolMinor: 1,
			URL:           mustParseURL("/foo"),
			Method:        "GET",
		},
		`GET /foo HTTP/1.1

`)
}

func TestHeaders(t *testing.T) {
	compare(t,
		&web.Request{
			ProtocolMajor: 1,
			ProtocolMinor: 1,
			URL:           mustParseURL("/foo"),
			Method:        "GET",
			Headers: map[string][]string{
				"Content-Type": []string{"text/html"},
                "Cookie": []string{"hello=world", "foo=bar"},
            },
		},
		`GeT /foo HTTP/1.1
Content-Type: text/html
Cookie: hello=world
Cookie: foo=bar

`)
}

func TestContinuationLine(t *testing.T) {
	compare(t,
		&web.Request{
			ProtocolMajor: 1,
			ProtocolMinor: 1,
			URL:           mustParseURL("/foo"),
			Method:        "GET",
			Headers: map[string][]string{
				"Content-Type": []string{"text/html"},
                "Cookie": []string{"hello=world, foo=bar"},
            },
		},
		`GeT /foo HTTP/1.1
Cookie: hello=world,
 foo=bar
Content-Type: text/html

`)
}

