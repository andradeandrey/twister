package main

import (
	"log"
	"flag"
	"fmt"
	"template"
	"github.com/garyburd/twister/web"
	"github.com/garyburd/twister/server"
)

func homeHandler(req *web.Request) {
	req.ParseForm()
	fmt.Println("hello", req.Param)
	homeTempl.Execute(req, req.Respond(web.StatusOK, web.HeaderContentType, "text/html"))
}

var homeTempl = template.MustParse(homeStr, nil)

const homeStr = `
<html>
<head>
<title>Request</title>
<style type="text/css">
.d {.meta-left}
    margin-left: 1em;
{.meta-right}
</style>
</head>
<body>
<a href="/">home</a><br>
<a href="/a/blorg">/a/blorg</a><br>
<a href="/a/foo?b=bar&amp;c=quux">/a/foo?b=bar&amp;c=quux</a><br>
<a href="/a/blorg/">/a/blorg/</a><br>
<a href="/b/foo/c/bar">/b/foo/c/bar</a><br>
<a href="/b/foo/c/bar/">/b/foo/c/bar/</a><br>
<form method="post" action="/c"><input type=text value="hello" name=b><input type="submit"></form></br>
<table>
<tr><th align="left" valign="top">Method</th><td>{Method}</td></tr>
<tr><th align="left" valign="top">URL</th><td>{URL}</td></tr>
<tr><th align="left" valign="top">ProtocolVersion</th><td>{ProtocolVersion}</td></tr>
<tr><th align="left" valign="top">Param</th><td>{Param}</td></tr>
<tr><th align="left" valign="top">Host</th><td>{Host}</td></tr>
<tr><th align="left" valign="top">ContentType</th><td>{ContentType}</td></tr>
<tr><th align="left" valign="top">ContentLength</th><td>{ContentLength}</td></tr>
<tr><th align="left" valign="top">Header</th><td>{Header}</td></tr>
</table>
</body>
</html> `


func main() {
	flag.Parse()
	r := web.NewRouter()
	r.Register("/", "GET", homeHandler)
	r.Register("/a/<a>/", "GET", homeHandler)
	r.Register("/b/<b>/c/<c>", "GET", homeHandler)
	r.Register("/c", "POST", homeHandler)

	hr := web.NewHostRouter(nil)
	hr.Register("www.example.com", r)

	err := server.ListenAndServe(":8080", hr)
	if err != nil {
		log.Exit("ListenAndServe:", err)
	}
}
