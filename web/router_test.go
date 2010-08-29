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
	"testing"
)

type rhandler string

func (rhandler) ServeHTTP(conn *http.Conn, req *http.Request) {}

func TestRouter(t *testing.T) {
	r := NewRouter(nil)
	r.Register("/", "GET", rhandler("home-get"))
	r.Register("/a", "GET", rhandler("a-get"), "*", rhandler("a-*"))
	r.Register("/b", "GET", rhandler("b-get"), "POST", rhandler("b-post"))
	r.Register("/c", "*", rhandler("c-*"))

	expectHandler := func(method string, path string, expectedName string, names []string, values []string) {
		handler, names, values := r.find(path, method)
		rhandler, ok := handler.(rhandler)
		if !ok {
			t.Errorf("Unexpected handler type for %s %s", method, path)
		}
		actualName := string(rhandler)
		if actualName != expectedName {
			t.Errorf("Unexpected handler for %s %s, actual %s expected %s", method, path, actualName, expectedName)
		}
	}

	expectError := func(method string, path string, statusCode int) {
		handler, _, _ := r.find(path, method)
		re, ok := handler.(*routerError)
		if !ok {
			t.Errorf("Unexpected handler type for %s %s", method, path)
		}
		if re.statusCode != statusCode {
			t.Errorf("Unexpected status for %s %s, actual %d expected %d", method, path, re.statusCode, statusCode)
		}
	}

	expectError("GET", "/Bogus/Path", 404)
	expectError("POST", "/Bogus/Path", 404)

	expectHandler("GET", "/", "home-get", nil, nil)
	expectHandler("HEAD", "/", "home-get", nil, nil)
	expectError("POST", "/", 405)

	expectHandler("GET", "/a", "a-get", nil, nil)
	expectHandler("HEAD", "/a", "a-get", nil, nil)
	expectHandler("POST", "/a", "a-*", nil, nil)

	expectHandler("GET", "/b", "b-get", nil, nil)
	expectHandler("HEAD", "/b", "b-get", nil, nil)
	expectHandler("POST", "/b", "b-post", nil, nil)
	expectError("PUT", "/b", 405)

	expectHandler("GET", "/c", "c-*", nil, nil)
	expectHandler("HEAD", "/c", "c-*", nil, nil)
}
