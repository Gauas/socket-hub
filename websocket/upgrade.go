package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func Upgrade(w http.ResponseWriter, r *http.Request) (net.Conn, *bufio.ReadWriter, bool) {
	if !Valid(r) {
		return nil, nil, false
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, false
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, false
	}

	_, _ = fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\n")
	_, _ = fmt.Fprintf(rw, "Upgrade: websocket\r\n")
	_, _ = fmt.Fprintf(rw, "Connection: Upgrade\r\n")
	_, _ = fmt.Fprintf(rw, "Sec-WebSocket-Accept: %s\r\n\r\n", accept(r.Header.Get("Sec-WebSocket-Key")))
	_ = rw.Flush()

	return conn, rw, true
}

func Valid(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

func accept(key string) string {
	hash := sha1.Sum([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(hash[:])
}
