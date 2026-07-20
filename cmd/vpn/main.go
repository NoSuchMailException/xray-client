package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/NoSuchMailException/xray-client/internal/inbound/socks5"
	"github.com/NoSuchMailException/xray-client/internal/outbound/xray"
	"github.com/NoSuchMailException/xray-client/internal/relay"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	inbound := socks5.NewServer("127.0.0.1:1080")
	outbound := xray.NewClient()
	relay := relay.NewRelay(inbound, outbound)

	go func() {
		slog.Info("[main] inbound listen to 127.0.0.1:1080")
		if err := inbound.ListenAndServe(ctx); err != nil && err != context.Canceled {
			slog.Error("[main] listen and serve error", "Error", err)
		}
	}()

	if err := relay.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("[main] relay error", "Error", err)
	}
}
