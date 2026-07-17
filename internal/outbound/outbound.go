// Package outbound defines the interface for dialing outgoing connections.
package outbound

import (
	"context"
	"net"
)

// Outbound dials outgoing connections to the destination through VPS.
type Outbound interface {
	// Connect dials the target address and returns an open connection.
	// The connection is closed when ctx is cancelled.
	Connect(ctx context.Context, target string) (net.Conn, error)
}
