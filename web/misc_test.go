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
	"testing"
	"reflect"
)

type ParseCookieValuesTest struct {
	values []string
	m      StringsMap
}

var ParseCookieValuesTests = []ParseCookieValuesTest{
	ParseCookieValuesTest{[]string{"a=b"}, StringsMap{"a": []string{"b"}}},
	ParseCookieValuesTest{[]string{"a=b; c"}, StringsMap{"a": []string{"b"}}},
	ParseCookieValuesTest{[]string{"a=b; =c"}, StringsMap{"a": []string{"b"}}},
	ParseCookieValuesTest{[]string{"a=b; ; "}, StringsMap{"a": []string{"b"}}},
	ParseCookieValuesTest{[]string{"a=b; c=d"}, StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
	ParseCookieValuesTest{[]string{"a=b; c=d"}, StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
	ParseCookieValuesTest{[]string{"a=b;c=d"}, StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
	ParseCookieValuesTest{[]string{" a=b;c=d "}, StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
	ParseCookieValuesTest{[]string{"a=b", "c=d"}, StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
}

func TestParseCookieValues(t *testing.T) {
	for _, pt := range ParseCookieValuesTests {
		m := parseCookieValues(pt.values)
		if !reflect.DeepEqual(pt.m, m) {
			t.Errorf("values=%s,\nexpected %q\nactual   %q", pt.values, pt.m, m)
		}
	}
}

type ParseUrlEncodedFormTest struct {
	s string
	m StringsMap
}

var ParseUrlEncodedFormTests = []ParseUrlEncodedFormTest{
	ParseUrlEncodedFormTest{"a=", StringsMap{"a": []string{""}}},
	ParseUrlEncodedFormTest{"a=b", StringsMap{"a": []string{"b"}}},
	ParseUrlEncodedFormTest{"a=b&c=d", StringsMap{"a": []string{"b"}, "c": []string{"d"}}},
	ParseUrlEncodedFormTest{"a=b&a=c", StringsMap{"a": []string{"b", "c"}}},
	ParseUrlEncodedFormTest{"a=Hello%20World", StringsMap{"a": []string{"Hello World"}}},
}

func TestParseUrlEncodedForm(t *testing.T) {
	for _, pt := range ParseUrlEncodedFormTests {
		p := []byte(pt.s)
		m := make(StringsMap)
		parseUrlEncodedFormBytes(p, m)
		if !reflect.DeepEqual(pt.m, m) {
			t.Errorf("form=%s,\nexpected %q\nactual   %q", pt.s, pt.m, m)
		}
	}
}
