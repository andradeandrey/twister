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
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strings"
)

type WebSocketConn struct {
	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer
}

func (ws *WebSocketConn) Close() os.Error {
	return ws.conn.Close()
}

func (ws *WebSocketConn) Receive() ([]byte, os.Error) {
    // Support text framing for now. Revisit after browsers support framing
    // described in later specs.
    c, err := ws.br.ReadByte()
	if err != nil {
		return nil, err
	}
	if c != 0 {
		return nil, os.NewError("twister.websocket: unexpected framing.")
	}
	p, err := ws.br.ReadSlice(0xff)
	if err != nil {
		return nil, err
	}
	return p[:len(p)-1], nil
}

func (ws *WebSocketConn) Send(p []byte) os.Error {
    // Support text framing for now. Revisit after browsers support framing
    // described in later specs.
    ws.bw.WriteByte(0)
	ws.bw.Write(p)
	ws.bw.WriteByte(0xff)
	return ws.bw.Flush()
}

// webSocketKey returns the key bytes from the specified websocket key header.
func webSocketKey(req *Request, name string) (key []byte, err os.Error) {
	s, found := req.Header.Get(name)
	if !found {
		return key, os.NewError("twister.websocket: missing key")
	}
	var n uint32 // number formed from decimal digits in key
	var d uint32 // number of spaces in key
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == ' ' {
			d += 1
		} else if '0' <= b && b <= '9' {
			n = n*10 + uint32(b) - '0'
		}
	}
	if d == 0 || n%d != 0 {
		return nil, os.NewError("twister.websocket: bad key")
	}
	key = make([]byte, 4)
	binary.BigEndian.PutUint32(key, n/d)
	return key, nil
}

func NewWebSocketConn(req *Request) (ws *WebSocketConn, err os.Error) {

	conn, buffered, err := req.Responder.Hijack()
	if err != nil {
		panic("twister.websocket: hijack failed")
		return nil, err
	}

	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	var r io.Reader
	if len(buffered) > 0 {
		r = io.MultiReader(bytes.NewBuffer(buffered), conn)
	} else {
		r = conn
	}
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(conn)

	if req.Method != "GET" {
		return nil, os.NewError("twister.websocket: bad request method")
	}

	origin, found := req.Header.Get(HeaderOrigin)
	if !found {
		return nil, os.NewError("twister.websocket: origin missing")
	}

	connection := strings.ToLower(req.Header.GetDef(HeaderConnection, ""))
	if connection != "upgrade" {
		return nil, os.NewError("twister.websocket: connection header missing or wrong value")
	}

	upgrade := strings.ToLower(req.Header.GetDef(HeaderUpgrade, ""))
	if upgrade != "websocket" {
		return nil, os.NewError("twister.websocket: upgrade header missing or wrong value")
	}

	key1, err := webSocketKey(req, HeaderSecWebSocketKey1)
	if err != nil {
		return nil, err
	}

	key2, err := webSocketKey(req, HeaderSecWebSocketKey2)
	if err != nil {
		return nil, err
	}

	key3 := make([]byte, 8)
	if _, err := io.ReadFull(br, key3); err != nil {
		return nil, err
	}

	h := md5.New()
	h.Write(key1)
	h.Write(key2)
	h.Write(key3)
	response := h.Sum()

	// TODO: handle tls
	location := "ws://" + req.Host + req.URL.RawPath
	protocol := req.Header.GetDef(HeaderSecWebSocketProtocol, "")

	bw.WriteString("HTTP/1.1 101 WebSocket Protocol Handshake")
	bw.WriteString("\r\nUpgrade: WebSocket")
	bw.WriteString("\r\nConnection: Upgrade")
	bw.WriteString("\r\nSec-WebSocket-Location: ")
	bw.WriteString(location)
	bw.WriteString("\r\nSec-WebSocket-Origin: ")
	bw.WriteString(origin)
	if len(protocol) > 0 {
		bw.WriteString("\r\nSec-WebSocket-Protocol: ")
		bw.WriteString(protocol)
	}
	bw.WriteString("\r\n\r\n")
	bw.Write(response)

	if err := bw.Flush(); err != nil {
		return nil, err
	}

	ws = &WebSocketConn{conn, br, bw}
	conn = nil
	return ws, nil
}
