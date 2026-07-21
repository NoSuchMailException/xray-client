package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/NoSuchMailException/xray-client/internal/config"
	inboundPkg "github.com/NoSuchMailException/xray-client/internal/inbound"
	"github.com/NoSuchMailException/xray-client/internal/inbound/socks5"
	"github.com/NoSuchMailException/xray-client/internal/outbound/xray"
	relayPkg "github.com/NoSuchMailException/xray-client/internal/relay"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("[config] load config", "err", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	address := fmt.Sprintf("%s:%d", cfg.Inbound.Listen, cfg.Inbound.Port)
	inbound := socks5.NewServer(address)
	outbound := xray.NewClient(cfg.Outbound)
	relay := relayPkg.NewRelay(inbound, outbound)

	go func() {
		slog.Info("[inbound] listening", "addr", address)
		if err := inbound.ListenAndServe(ctx); err != nil && err != context.Canceled {
			slog.Error("[inbound] listen and serve", "err", err)
		}
	}()

	if err := relay.Run(ctx); err != nil &&
		err != context.Canceled &&
		!errors.Is(err, inboundPkg.ErrClosed) {
		slog.Error("[relay] stopped", "err", err)
	}
}
