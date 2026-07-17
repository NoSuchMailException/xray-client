// Package relay connects inbound and outbound, proxying data between them.
package relay

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/NoSuchMailException/xray-client/internal/inbound"
	"github.com/NoSuchMailException/xray-client/internal/outbound"
)

// Relay bridges an inbound listener and an outbound dialer.
// Each accepted Request is proxied to its target in a separate goroutine.
type Relay struct {
	Inbound  inbound.Inbound
	Outbound outbound.Outbound
}

// NewRelay creates a new Relay with the given inbound and outbound.
func NewRelay(in inbound.Inbound, out outbound.Outbound) *Relay {
	return &Relay{
		Inbound:  in,
		Outbound: out,
	}
}

// Run accepts requests from inbound in a loop and proxies each one
// through outbound until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := r.Inbound.Accept()
		if err != nil {
			return fmt.Errorf("inbound accept: %w", err)
		}

		target := req.Target
		inboundConn := req.Conn

		outboundConn, err := r.Outbound.Connect(ctx, target)
		if err != nil {
			return fmt.Errorf("outbound connect: %w", err)
		}

		go func(localConn, vpsConn net.Conn) {
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				io.Copy(vpsConn, localConn)
			}()

			go func() {
				defer wg.Done()
				io.Copy(localConn, vpsConn)
			}()

			wg.Wait()
			localConn.Close()
			vpsConn.Close()
		}(inboundConn, outboundConn)
	}
}
