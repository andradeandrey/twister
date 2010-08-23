package main

import (
    "http"
    "log"
    "template"
    "github.com/garyburd/twister/web"
)

func homeHandler(c *http.Conn, req *http.Request) {
    homeTempl.Execute(req, c)
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
<a href="/a/blorg/">/a/blorg/</a><br>
<a href="/b/foo/c/bar">/b/foo/c/bar</a><br>
<a href="/b/foo/c/bar/">/b/foo/c/bar/</a><br>
<hr>
{Form}
</body>
</html> `

func main() {
    r  := web.NewRouter()
    r.Register("/", "GET",  homeHandler)
    r.Register("/a/<a>/", "GET", homeHandler)
    r.Register("/b/<a>/c/<b>", "GET", homeHandler)
    err := http.ListenAndServe(":8080", r)
    if err != nil {
        log.Exit("ListenAndServe:", err)
    }
}
