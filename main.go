package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/andy-ta/andydb/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := app.ServerConfig{
		Addr:            ":42069",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	server := &app.App{}
	server.Initialize()
	if err := server.Run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server terminated: %v", err)
	}
}
