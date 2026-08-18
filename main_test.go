package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/andy-ta/andydb/app"
)

func restoreMainHooks(t *testing.T) {
	t.Helper()
	prevLookup := lookupEnv
	prevArg0 := arg0
	prevFatalf := fatalf
	prevStart := startServer
	t.Cleanup(func() {
		lookupEnv = prevLookup
		arg0 = prevArg0
		fatalf = prevFatalf
		startServer = prevStart
	})
}

func TestMainSuccess(t *testing.T) {
	restoreMainHooks(t)
	lookupEnv = func(string) string { return "prod" }
	startServer = func(context.Context, app.ServerConfig) error { return nil }
	main()
}

func TestMainCanceled(t *testing.T) {
	restoreMainHooks(t)
	lookupEnv = func(string) string { return "prod" }
	startServer = func(context.Context, app.ServerConfig) error { return context.Canceled }
	main()
}

func TestMainFatalOnError(t *testing.T) {
	restoreMainHooks(t)
	logged := false
	fatalf = func(string, ...interface{}) { logged = true }
	lookupEnv = func(string) string { return "prod" }
	startServer = func(context.Context, app.ServerConfig) error { return errors.New("boom") }
	main()
	if !logged {
		t.Fatal("expected fatalf on unexpected server error")
	}
}

func TestMainDevEnv(t *testing.T) {
	restoreMainHooks(t)
	lookupEnv = func(key string) string {
		if key == "ANDYDB_ENV" {
			return "dev"
		}
		return ""
	}
	var got app.ServerConfig
	startServer = func(_ context.Context, cfg app.ServerConfig) error {
		got = cfg
		return nil
	}
	main()
	if !got.DevMode {
		t.Fatal("expected ANDYDB_ENV=dev to enable dev mode")
	}
}

func TestMainEmptyEnvUnderGoRun(t *testing.T) {
	restoreMainHooks(t)
	lookupEnv = func(string) string { return "" }
	arg0 = func() string { return "/tmp/go-build123/exe/andydb" }
	var got app.ServerConfig
	startServer = func(_ context.Context, cfg app.ServerConfig) error {
		got = cfg
		return nil
	}
	main()
	if !got.DevMode {
		t.Fatal("expected go run heuristic to enable dev mode")
	}
}

func TestMainEmptyEnvCompiledBinary(t *testing.T) {
	restoreMainHooks(t)
	lookupEnv = func(string) string { return "" }
	arg0 = func() string { return "/usr/local/bin/andydb" }
	var got app.ServerConfig
	startServer = func(_ context.Context, cfg app.ServerConfig) error {
		got = cfg
		return nil
	}
	main()
	if got.DevMode {
		t.Fatal("expected compiled binary without env to stay in production mode")
	}
}

func TestRunningUnderGoRun(t *testing.T) {
	restoreMainHooks(t)
	arg0 = func() string { return "/tmp/go-build/exe/andydb" }
	if !runningUnderGoRun() {
		t.Fatal("expected go-build path to match")
	}
	arg0 = func() string { return "/usr/bin/andydb" }
	if runningUnderGoRun() {
		t.Fatal("expected non go-build path to miss")
	}
}

func TestDefaultArg0(t *testing.T) {
	if strings.TrimSpace(arg0()) == "" {
		t.Fatal("expected default arg0 to return the process path")
	}
}

func TestStartServerDefault(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- startServer(ctx, app.ServerConfig{
			Addr:            addr,
			ReadTimeout:     time.Second,
			WriteTimeout:    time.Second,
			IdleTimeout:     time.Second,
			ShutdownTimeout: 2 * time.Second,
		})
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, getErr := http.Get("http://" + addr + "/api/missing")
		if getErr == nil {
			resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("default startServer never became ready: %v", getErr)
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case runErr := <-errCh:
		if !errors.Is(runErr, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("startServer did not return after cancel")
	}
}
