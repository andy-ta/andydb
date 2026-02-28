package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/andy-ta/andydb/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envMode := os.Getenv("ANDYDB_ENV")
	devMode := strings.EqualFold(envMode, "dev")
	if envMode == "" && runningUnderGoRun() {
		devMode = true
	}
	cfg := app.ServerConfig{
		Addr:            ":42069",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		DevMode:         devMode,
	}

	server := &app.App{}
	server.Initialize()
	if err := server.Run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server terminated: %v", err)
	}
}

func runningUnderGoRun() bool {
	return strings.Contains(os.Args[0], "go-build")
}
