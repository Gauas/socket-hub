package http

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type socketConn struct {
	conn    net.Conn
	timeout time.Duration
}

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel is required"})
		return
	}
	if !isWebSocket(r) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "websocket upgrade is required"})
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "websocket unsupported"})
		return
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	accept := acceptKey(key)
	_, _ = fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\n")
	_, _ = fmt.Fprintf(rw, "Upgrade: websocket\r\n")
	_, _ = fmt.Fprintf(rw, "Connection: Upgrade\r\n")
	_, _ = fmt.Fprintf(rw, "Sec-WebSocket-Accept: %s\r\n\r\n", accept)
	_ = rw.Flush()

	socket := &socketConn{conn: conn, timeout: s.cfg.WriteTimeout}
	client := s.hub.Subscribe(channel, socket)
	defer s.hub.Unsubscribe(client)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		defer cancel()
		_ = readFrames(rw.Reader)
	}()

	client.Write(ctx)
}

func isWebSocket(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

func acceptKey(key string) string {
	hash := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (c *socketConn) Write(ctx context.Context, payload []byte) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.timeout)
	}
	_ = c.conn.SetWriteDeadline(deadline)
	return writeText(c.conn, payload)
}

func (c *socketConn) Close() error {
	return c.conn.Close()
}

func writeText(w io.Writer, payload []byte) error {
	header := []byte{0x81}
	size := len(payload)

	switch {
	case size < 126:
		header = append(header, byte(size))
	case size <= 65535:
		header = append(header, 126, byte(size>>8), byte(size))
	default:
		header = append(header, 127)
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(size))
		header = append(header, length...)
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func readFrames(r *bufio.Reader) error {
	for {
		header, err := r.ReadByte()
		if err != nil {
			return err
		}
		opcode := header & 0x0f

		sizeByte, err := r.ReadByte()
		if err != nil {
			return err
		}
		masked := sizeByte&0x80 != 0
		size := int64(sizeByte & 0x7f)

		switch size {
		case 126:
			var ext [2]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				return err
			}
			size = int64(binary.BigEndian.Uint16(ext[:]))
		case 127:
			var ext [8]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				return err
			}
			size = int64(binary.BigEndian.Uint64(ext[:]))
		}

		if masked {
			var mask [4]byte
			if _, err := io.ReadFull(r, mask[:]); err != nil {
				return err
			}
		}
		if size > 0 {
			if _, err := io.CopyN(io.Discard, r, size); err != nil {
				return err
			}
		}
		if opcode == 0x8 {
			return nil
		}
	}
}
