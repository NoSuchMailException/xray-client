// Package inbound defines the interface for accepting incoming connections.
// Each implementation parses its own protocol and returns a unified Request.
package inbound

import (
	"context"
	"net"
)

// Request represents an incoming connection from the client
// with the destination address it wants to reach.
type Request struct {
	// Conn is the connection from the browser to our client.
	Conn net.Conn
	// Target is the destination address, for example "github.com:443".
	Target string
}

// Inbound listens to incoming traffic and parses it into Requests.
// SOCKS5 and TUN are different implementations of this interface.
type Inbound interface {
	// ListenAndServe starts a listener on configured address and port.
	// Blocks until context is cancelled.
	ListenAndServe(ctx context.Context) error
	// Accept returns the next ready Request.
	// Blocks until a new connection arrives.
	Accept() (*Request, error)
}
