package websocket

import (
	"context"
	"net"
	"time"
)

type Connection struct {
	conn    net.Conn
	timeout time.Duration
}

func New(conn net.Conn, timeout time.Duration) *Connection {
	return &Connection{conn: conn, timeout: timeout}
}

func (c *Connection) Write(ctx context.Context, payload []byte) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.timeout)
	}
	_ = c.conn.SetWriteDeadline(deadline)
	return Write(c.conn, payload)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
