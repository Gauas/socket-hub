package websocket

import (
	"context"
	"net"
	"time"
)

type Conn struct {
	conn    net.Conn
	timeout time.Duration
}

func New(conn net.Conn, timeout time.Duration) *Conn {
	return &Conn{conn: conn, timeout: timeout}
}

func (c *Conn) Write(ctx context.Context, payload []byte) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.timeout)
	}
	_ = c.conn.SetWriteDeadline(deadline)
	return Write(c.conn, payload)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
