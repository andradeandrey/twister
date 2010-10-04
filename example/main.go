package main

import (
	"log"
	"flag"
	"github.com/garyburd/twister/web"
	"github.com/garyburd/twister/server"
	"template"
)

func homeHandler(req *web.Request) {
	homeTempl.Execute(req,
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html"))
}

func main() {
	flag.Parse()
	h := web.SetErrorHandler(coreErrorHandler,
		web.ProcessForm(10000, true, web.NewHostRouter(nil).
			Register("www.example.com", web.NewRouter().
			Register("/", "GET", homeHandler).
			Register("/chat", "GET", chatFrameHandler).
			Register("/chat/ws", "GET", chatWsHandler).
			Register("/core/", "GET", coreHandler).
			Register("/core/a/<a>/", "GET", coreHandler).
			Register("/core/b/<b>/c/<c>", "GET", coreHandler).
			Register("/core/c", "POST", coreHandler))))

	err := server.ListenAndServe("localhost:8080", ":8080", h)
	if err != nil {
		log.Exit("ListenAndServe:", err)
	}
}

var homeTempl = template.MustParse(homeStr, nil)

const homeStr = `
<html>
<head>
</head>
<body>
<ul>
<li><a href="/core">Core functionality</a>
<li><a href="/chat">Chat using WebSockets</a>
</ul>
</body>
</html>`
