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
	"crypto/rand"
	"encoding/hex"
)

type respondFilter struct {
	Responder
	filter func(status int, header StringsMap) (int, StringsMap)
}

func (rf *respondFilter) Respond(status int, header StringsMap) ResponseBody {
	return rf.Responder.Respond(rf.filter(status, header))
}

// FilterRespond replaces the request's responder with one that filters the
// arguments to Respond through the supplied filter. This function is intended
// to be used by middleware.
func FilterRespond(req *Request, filter func(status int, header StringsMap) (int, StringsMap)) {
	req.Responder = &respondFilter{req.Responder, filter}
}

// SetErrorHandler returns a handler that sets the request's error handler to the supplied handler.
func SetErrorHandler(errorHandler func(req *Request, status int, message string), handler Handler) Handler {
    return HandlerFunc(func(req *Request) {
        req.ErrorHandler = errorHandler
        handler.ServeWeb(req)
    })
}

const (
	XSRFCookieName = "xsrf"
	XSRFParamName  = "xsrf"
)

// ProcessForm returns a handler that checks the request body length, parses
// url encoded forms and optionaly checks for XRSF.
func ProcessForm(maxRequestBodyLen int, checkXSRF bool, handler Handler) Handler {
	return HandlerFunc(func(req *Request) {

		if req.ContentLength > maxRequestBodyLen {
			status := StatusRequestEntityTooLarge
			if _, found := req.Header.Get(HeaderExpect); found {
				status = StatusExpectationFailed
			}
			req.Error(status, "Request entity too large.")
			return
		}

		if err := req.ParseForm(); err != nil {
			req.Error(StatusBadRequest, "Error reading or parsing form.")
			return
		}

		if checkXSRF {
            const tokenLen = 8
			token, found := req.Cookie.Get(XSRFCookieName)

            // Create new XSRF token?
            if !found || len(token) != tokenLen {
				p := make([]byte, tokenLen/2)
				_, err := rand.Reader.Read(p)
				if err != nil {
					panic("twister: rand read failed")
				}
				token = hex.EncodeToString(p)
				c := Cookie{
					Name:     XSRFCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
				}
				value := c.String()
				FilterRespond(req, func(status int, header StringsMap) (int, StringsMap) {
					header.Append(HeaderSetCookie, value)
					return status, header
				})
			}

            if token != req.Param.GetDef(XSRFParamName, "") {
				req.Param.Set(XSRFParamName, token)
			    if (req.Method == "POST" || req.Method == "PUT") {
				    req.Error(StatusNotFound, "Bad token")
				    return
				}
            }
		}

		handler.ServeWeb(req)
	})
}
