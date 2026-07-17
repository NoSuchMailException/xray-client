package socks5

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/NoSuchMailException/xray-client/internal/inbound"
)

var _ inbound.Inbound = (*Server)(nil)

type Server struct {
	addr    string
	channel chan *inbound.Request
}

func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		channel: make(chan *inbound.Request),
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		slog.Error("[socks5] listener connect error", "Error", err)
		return fmt.Errorf("listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("[socks5] get connection error", "Error", err)
			return err
		}

		// Go 1.22+: loop variable is copied per iteration, no shadowing needed
		go func() {
			target, err := handshakeSocks5(conn)
			if err != nil {
				slog.Error("[socks5] handshake error", "Error", err)
				conn.Close()
				return
			}
			s.channel <- &inbound.Request{
				Conn:   conn,
				Target: target,
			}
		}()
	}
}

func (s *Server) Accept() (*inbound.Request, error) {
	return <-s.channel, nil
}

func handshakeSocks5(conn net.Conn) (string, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", fmt.Errorf("greeting read: %w", err)
	}
	if header[0] != 0x05 {
		return "", fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}

	methods := make([]byte, header[1])
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", fmt.Errorf("methods read: %w", err)
	}

	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return "", fmt.Errorf("greeting write: %w", err)
	}

	reqHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, reqHeader); err != nil {
		return "", fmt.Errorf("request header read: %w", err)
	}
	if reqHeader[1] != 0x01 {
		return "", fmt.Errorf("unsupported command: %d (only TCP CONNECT)", reqHeader[1])
	}

	var host string
	atyp := reqHeader[3]

	switch atyp {
	case 0x01:
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", fmt.Errorf("ipv4 read: %w", err)
		}
		host = net.IP(ip).String()
	case 0x03:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", fmt.Errorf("domain len read: %w", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", fmt.Errorf("domain read: %w", err)
		}
		host = string(domain)
	case 0x04:
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", fmt.Errorf("ipv6 read: %w", err)
		}
		host = net.IP(ip).String()
	default:
		return "", fmt.Errorf("unknown address type: %d", atyp)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", fmt.Errorf("port read: %w", err)
	}
	port := binary.BigEndian.Uint16(portBuf)

	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return "", fmt.Errorf("success response write: %w", err)
	}
	return fmt.Sprintf("%s:%d", host, port), nil
}
