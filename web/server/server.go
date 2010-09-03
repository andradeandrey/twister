package server

import (
	"bufio"
	"os"
	"bytes"
	"github.com/garyburd/twister/web"
	"http"
	"regexp"
	"strconv"
)

func skipBytes(p []byte, f func(byte) bool) int {
	i := 0
    for ; i < len(p); i++ {
		if !f(byte(p[i])) {
			break
		}
	}
	return i
}

func skipBytesRev(p []byte, f func(byte) bool) int {
	var i int
	for i = len(p); i > 0; i-- {
		if !f(p[i-1]) {
			break
		}
	}
	return i
}

func readLine(b *bufio.Reader) ([]byte, os.Error) {
	p, err := b.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	n := len(p) - 1
	if n > 1 && p[n-1] == '\r' {
		n = n - 1
	}
	return p[0:n], nil
}

var requestLineRegexp = regexp.MustCompile("^([_A-Za-z0-9]+) ([^ ]+) HTTP/([0-9]+)\\.([0-9]+)[ \t]*$")

func parseRequest(b *bufio.Reader) (*web.Request, os.Error) {

	req := web.Request{}

	p, err := readLine(b)
	if err != nil {
		return nil, err
	}

	m := requestLineRegexp.FindSubmatch(p)
	if m == nil {
		return nil, os.ErrorString("malformed request line")
	}

	req.Method = string(bytes.ToUpper(m[1]))

	req.ProtocolMajor, err = strconv.Atoi(string(m[3]))
	if err != nil {
		return nil, os.ErrorString("bad major version")
	}

	req.ProtocolMinor, err = strconv.Atoi(string(m[4]))
	if err != nil {
		return nil, os.ErrorString("bad minor version")
	}

	req.URL, err = http.ParseURL(string(m[2]))
	if err != nil {
		return nil, os.ErrorString("bad url")
	}

	req.Headers, err = parseHeaders(b)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func parseHeaders(b *bufio.Reader) (web.Headers, os.Error) {

	const (
		// Max size for header line
		maxLineSize = 4096
		// Max size for header value
		maxValueSize = 4096
		// Maximum number of headers 
		maxHeaderCount = 256
	)

	headers := web.Headers(make(map[string][]string))
	lastKey := ""
	headerCount := 0

	for {
		p, err := readLine(b)
		if err != nil {
			return nil, err
		}

		// don't allow huge headers
		if len(p) > maxLineSize {
			return nil, os.ErrorString("header line too long")
		}

		// End of headers?
		if len(p) == 0 {
			return headers, nil
		}

		if web.IsSpaceByte(p[0]) {

			// Continuation of previous header line

			if lastKey == "" {
				return nil, os.ErrorString("continuation before first header")
			}

			values := headers[lastKey]
			value := values[len(values)-1]
			value = value + string(p)
			if len(value) > maxValueSize {
				return nil, os.ErrorString("header value too long")
			}
			values[len(values)-1] = value

		} else {

			// New header

			headerCount = headerCount + 1
			if headerCount > maxHeaderCount {
				return nil, os.ErrorString("too many headers")
			}

			// Key
			i := skipBytes(p, web.IsTokenByte)
			if i < 1 {
				return nil, os.ErrorString("missing header key")
			}
			web.CanonicalizeHeaderKeyBytes(p[0:i])
			key := string(p[0:i])
			p = p[i:]

			// Skip WS
			p = p[skipBytes(p, web.IsSpaceByte):]

			// Colon
			if p[0] != ':' {
				return nil, os.ErrorString("missing :")
			}
			p = p[1:]

			// Skip WS
			p = p[skipBytes(p, web.IsSpaceByte):]

			// Trim trailing WS
			p = p[0:skipBytesRev(p, web.IsSpaceByte)]

			// Value
			value := string(p)
			headers.Add(key, value)
			lastKey = key
		}
	}
	return headers, nil
}
