package xray

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/NoSuchMailException/xray-client/internal/config"
	"github.com/NoSuchMailException/xray-client/internal/outbound"
	tls "github.com/refraction-networking/utls"
)

var _ outbound.Outbound = (*Client)(nil)

type Client struct {
	cfg config.OutboundConfig
}

func NewClient(cfg config.OutboundConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Connect(ctx context.Context, target string) (net.Conn, error) {

	serverPub, err := base64.RawURLEncoding.DecodeString(c.cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}

	shortID, err := hex.DecodeString(c.cfg.ShortID)
	if err != nil {
		return nil, fmt.Errorf("decode short id: %w", err)
	}

	sessionID, ePub, ePriv, err := buildSessionID(serverPub, shortID)
	if err != nil {
		return nil, fmt.Errorf("build session id: %w", err)
	}

	tcpConn, err := net.Dial("tcp", net.JoinHostPort(c.cfg.Address, fmt.Sprintf("%d", c.cfg.Port)))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	tlsUConn := tls.UClient(tcpConn, &tls.Config{
		ServerName:         c.cfg.ServerName,
		InsecureSkipVerify: true,
	}, tls.HelloFirefox_Auto)

	if err := tlsUConn.BuildHandshakeState(); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("build handshake state: %w", err)
	}

	tlsUConn.HandshakeState.Hello.SessionId = sessionID[:]

	for i, ks := range tlsUConn.HandshakeState.Hello.KeyShares {
		if ks.Group == tls.X25519 {
			tlsUConn.HandshakeState.Hello.KeyShares[i].Data = ePub[:]
			break
		}
	}

	x25519 := ecdh.X25519()
	ecdhPrivKey, err := x25519.NewPrivateKey(ePriv[:])
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("ecdh private key: %w", err)
	}

	tlsUConn.HandshakeState.State13.KeyShareKeys = &tls.KeySharePrivateKeys{
		Ecdhe: ecdhPrivKey,
	}

	if err := tlsUConn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("tls handshake: %w", err)
	}

	go func() {
		<-ctx.Done()
		tlsUConn.Close()
	}()

	return tlsUConn, nil
}
