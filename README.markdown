## Overview

Twister is a collection of useful utilities for writing web applications in
[Go](http://golang.org/). 

Twister includes the following features:

* Request routing using path and method.
* Request routing using host header

## Installation

1. [Install Go](http://golang.org/doc/install.html).
2. `goinstall github.com/garyburd/twister/web`

## Router Example

    import (
        "web"
        "http"
    )

    func main() {
        r  := web.NewRouter(nil)
        r.Register("/", "GET",  homeHandler)
        r.Register("/<page>", "GET", viewHandler)
        r.Register("/<page>/edit", "GET", editHandler, "POST", saveHandler)
        err := http.ListenAndServe(":8080", r)
        if err != nil {
            log.Exit("ListenAndServe:", err)
        }
    }

    func homeHandler(c *http.Conn, req *http.Request) {
        // display home page
    }

    func viewHandler(c *http.Conn, req *http.Request) {
        page := req.FormValue("page")
        // display page
    }

    func editHandler(c *http.Conn, req *http.Request) {
        page := req.FormValue("page")
        // display page edit form
    }

    func saveHandler(c *http.Conn, req *http.Request) {
        page := req.FormValue("page")
        // save page to db
    }

## About

Twister was written by [Gary Burd](http://gary.beagledreams.com/). The name
"Twister" was inspired by [Tornado](http://tornadoweb.org/").

