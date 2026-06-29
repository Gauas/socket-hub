package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gauas/socket-hub/config"
	"github.com/Gauas/socket-hub/realtime"
	"github.com/Gauas/socket-hub/server"
)

func main() {
	cfg := config.New()
	hub := realtime.New()
	server := server.New(cfg, hub)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
