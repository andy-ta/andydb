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

var (
	lookupEnv     = os.Getenv
	arg0          = func() string { return os.Args[0] }
	fatalf        = log.Fatalf
	notifyContext = signal.NotifyContext
	startServer   = func(ctx context.Context, cfg app.ServerConfig) error {
		server := &app.App{}
		server.Initialize()
		return server.Run(ctx, cfg)
	}
)

func main() {
	ctx, stop := notifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envMode := lookupEnv("ANDYDB_ENV")
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

	if err := startServer(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		fatalf("server terminated: %v", err)
	}
}

func runningUnderGoRun() bool {
	return strings.Contains(arg0(), "go-build")
}
