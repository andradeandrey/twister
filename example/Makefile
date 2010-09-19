run: clean goinstall example
	./example -hostOverride=www.example.com
 
include $(GOROOT)/src/Make.inc

TARG=example
GOFILES=\
    main.go\

include $(GOROOT)/src/Make.cmd

goinstall:
	goinstall github.com/garyburd/twister/server
	goinstall github.com/garyburd/twister/web

