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
<a href="/a/blorg/">/a/blorg/</a><br>
<a href="/b/foo/c/bar">/b/foo/c/bar</a><br>
<a href="/b/foo/c/bar/">/b/foo/c/bar/</a><br>
<hr>
{Form}
</body>
</html> `

func main() {
    flag.Parse()
    r  := web.NewRouter()
    r.Register("/", "GET",  homeHandler)
    r.Register("/a/<a>/", "GET", homeHandler)
    r.Register("/b/<a>/c/<b>", "GET", homeHandler)

    hr := web.NewHostRouter(nil)
    hr.Register("www.example.com", r)

    fmt.Println("a")
    err := server.ListenAndServe(":8080", hr)
    fmt.Println("b")
    if err != nil {
        log.Exit("ListenAndServe:", err)
    }
}
