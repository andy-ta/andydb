package app

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecovererReturnsJSONError(t *testing.T) {
	a := &App{}
	wrapped := a.recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got %q: %v", rec.Body.String(), err)
	}
	if body["error"] != "internal server error" {
		t.Fatalf("unexpected error payload: %#v", body)
	}
}

func TestRunGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	a := &App{}
	a.Initialize()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(ctx, ServerConfig{
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
			t.Fatalf("server never became ready: %v", getErr)
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
		t.Fatal("Run did not return after cancel")
	}
}
