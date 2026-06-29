package realtime

import "context"

type Conn interface {
	Write(context.Context, []byte) error
	Close() error
}

type Client struct {
	channel string
	conn    Conn
	send    chan []byte
	done    chan struct{}
}

func (c *Client) Write(ctx context.Context) {
	defer c.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, message); err != nil {
				return
			}
		case <-c.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
		_ = c.conn.Close()
	}
}
