package main

import (
	"template"
	"github.com/garyburd/twister/web"
)

func coreErrorHandler(req *web.Request, status int, message string) {
	coreTempl.Execute(map[string]interface{}{
		"req":     req,
		"status":  status,
		"message": message,
		"xsrf":    req.Param.GetDef(web.XSRFParamName, ""),
	},
		req.Respond(status, web.HeaderContentType, "text/html"))
}

func coreHandler(req *web.Request) {
	coreTempl.Execute(map[string]interface{}{
		"req":     req,
		"status":  web.StatusOK,
		"message": "ok",
		"xsrf":    req.Param.GetDef(web.XSRFParamName, ""),
	},
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html"))
}

var coreTempl = template.MustParse(coreStr, nil)

const coreStr = `
<html>
<head>
<title>Core</title>
</head>
<body>
<hr>
Status: {status} {message}
<hr>
<a href="/core">/core</a><br>
<a href="/core/a/blorg">/core/a/blorg</a><br>
<a href="/core/a/foo?b=bar&amp;c=quux">/core/a/foo?b=bar&amp;c=quux</a><br>
<a href="/core/a/blorg/">/core/a/blorg/</a><br>
<a href="/core/b/foo/c/bar">/core/b/foo/c/bar</a><br> 
<a href="/core/b/foo/c/bar/">/core/b/foo/c/bar/</a> (not found)<br>
<form method="post" action="/core/c"><input type="hidden" name="xsrf" value="{xsrf}"><input type=text value="hello" name=b><input type="submit"></form>
<form method="post" action="/core/c"><input type=text value="hello" name=b><input value="xsrf fail" type="submit"></form>
<hr>
{.section req}
<table>
<tr><th align="left" valign="top">RemoteAddr</th><td>{RemoteAddr}</td></tr>
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
