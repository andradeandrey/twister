package main

import (
	"log"
	"flag"
	"template"
	"github.com/garyburd/twister/web"
	"github.com/garyburd/twister/server"
)

func errorHandler(req *web.Request, status int, message string) {
	homeTempl.Execute(map[string]interface{}{
		"req":     req,
		"status":  status,
		"message": message,
		"xsrf":    req.Param.GetDef(web.XSRFParamName, ""),
	},
		req.Respond(status, web.HeaderContentType, "text/html"))
}


func homeHandler(req *web.Request) {
	homeTempl.Execute(map[string]interface{}{
		"req":     req,
		"status":  web.StatusOK,
		"message": "ok",
		"xsrf":    req.Param.GetDef(web.XSRFParamName, ""),
	},
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html"))
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
<a href="/b/foo/c/bar/">/b/foo/c/bar/</a> (not found)<br>
<form method="post" action="/c"><input type="hidden" name="xsrf" value="{xsrf}"><input type=text value="hello" name=b><input type="submit"></form>
<form method="post" action="/c"><input type=text value="hello" name=b><input value="xsrf fail" type="submit"></form>
<a href="/chat">chat</a>
<hr>
Status: {status} {message}
<hr>
{.section req}
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
{.end}
</body>
</html> `

func main() {
	flag.Parse()
	h := web.SetErrorHandler(errorHandler,
		web.ProcessForm(10000, true, web.NewHostRouter(nil).
			Register("www.example.com", web.NewRouter().
			Register("/chat", "GET", chatFrameHandler).
			Register("/chat/ws", "GET", chatWsHandler).
			Register("/", "GET", homeHandler).
			Register("/a/<a>/", "GET", homeHandler).
			Register("/b/<b>/c/<c>", "GET", homeHandler).
			Register("/c", "POST", homeHandler))))

	err := server.ListenAndServe(":8080", h)
	if err != nil {
		log.Exit("ListenAndServe:", err)
	}
}
