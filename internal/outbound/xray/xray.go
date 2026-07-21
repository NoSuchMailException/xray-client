package xray

import (
	"context"
	"fmt"
	"net"

	"github.com/NoSuchMailException/xray-client/internal/config"
	"github.com/NoSuchMailException/xray-client/internal/outbound"
)

var _ outbound.Outbound = (*Client)(nil)

type Client struct {
	cfg config.OutboundConfig
}

func NewClient(cfg config.OutboundConfig) *Client {
	return &Client{cfg: cfg}
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
