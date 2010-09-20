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
	"github.com/garyburd/twister/web"
)

type parseTest struct {
	name    string
	method  string
	url     string
	version int
	header  web.StringsMap
	s       string
}

var parseTests = []parseTest{
	parseTest{"no body", "", "", 0, web.NewStringsMap(), "GET /foo HTTP/1.1\r\n"},
	parseTest{"empty", "", "", 0, web.NewStringsMap(), " "},
	parseTest{"simple", "GET", "/foo", web.ProtocolVersion(1, 1), web.NewStringsMap(),
		`GET /foo HTTP/1.1

`},
	parseTest{"multihdr", "GET", "/foo", web.ProtocolVersion(1, 0), web.NewStringsMap(
		web.HeaderContentType, "text/html",
		web.HeaderCookie, "hello=world",
		web.HeaderCookie, "foo=bar"),
		`GET /foo HTTP/1.0
Content-Type: text/html
CoOkie: hello=world
Cookie: foo=bar

`},
	parseTest{"continuation", "GET", "/hello", web.ProtocolVersion(1, 1), web.NewStringsMap(
		web.HeaderContentType, "text/html",
		web.HeaderCookie, "hello=world, foo=bar"),
		`GET /hello HTTP/1.1
Cookie: hello=world,
 foo=bar
Content-Type: text/html

`},
}

func TestParse(t *testing.T) {
	for _, tt := range parseTests {
		b := bufio.NewReader(bytes.NewBufferString(tt.s))
		method, url, version, statusErr := parseRequestLine(b)
		header, headerErr := parseHeader(b)
		if tt.method == "" {
			if statusErr == nil && headerErr == nil {
				t.Errorf("%s: expected error", tt.name)
			}
		} else {
			if tt.method != method {
				t.Errorf("%s: method=%s, expected %s", tt.name, method, tt.method)
			}
			if tt.url != url {
				t.Errorf("%s: url=%s, expected %s", tt.name, url, tt.url)
			}
			if tt.version != version {
				t.Errorf("%s: version=%d, expected %d", tt.name, version, tt.version)
			}
			if !reflect.DeepEqual(tt.header, header) {
				t.Errorf("%s bad header\nexpected: %q\nactual:   %q", tt.name, tt.header, header)
			}
		}
	}
}
