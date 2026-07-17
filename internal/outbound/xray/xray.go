package xray

import (
	"context"
	"fmt"
	"net"

	"github.com/NoSuchMailException/xray-client/internal/outbound"
)

var _ outbound.Outbound = (*Client)(nil)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(ctx context.Context, target string) (net.Conn, error) {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return conn, nil
}
