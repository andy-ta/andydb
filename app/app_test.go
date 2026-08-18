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

func TestRequestLogger(t *testing.T) {
	a := &App{}
	called := false
	wrapped := a.requestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/logged", nil))
	if !called {
		t.Fatal("expected next handler to run")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestRecovererPassesThrough(t *testing.T) {
	a := &App{}
	wrapped := a.recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func restoreServerHooks(t *testing.T) {
	t.Helper()
	prevListen := listenAndServe
	prevShutdown := shutdownServer
	t.Cleanup(func() {
		listenAndServe = prevListen
		shutdownServer = prevShutdown
	})
}

func TestRunNilContextListenError(t *testing.T) {
	restoreServerHooks(t)
	listenAndServe = func(*http.Server) error { return errors.New("bind failed") }

	a := &App{}
	a.Initialize()
	err := a.Run(nil, ServerConfig{Addr: "127.0.0.1:0", DevMode: true})
	if err == nil || err.Error() != "bind failed" {
		t.Fatalf("expected bind failed, got %v", err)
	}
}

func TestRunListenServerClosed(t *testing.T) {
	restoreServerHooks(t)
	listenAndServe = func(*http.Server) error { return http.ErrServerClosed }

	a := &App{}
	a.Initialize()
	if err := a.Run(context.Background(), ServerConfig{Addr: "127.0.0.1:0"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunShutdownError(t *testing.T) {
	restoreServerHooks(t)
	started := make(chan struct{})
	released := make(chan struct{})
	listenAndServe = func(*http.Server) error {
		close(started)
		<-released
		return http.ErrServerClosed
	}
	shutdownServer = func(*http.Server, context.Context) error {
		return errors.New("shutdown failed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		a := &App{}
		a.Initialize()
		errCh <- a.Run(ctx, ServerConfig{Addr: "127.0.0.1:0", ShutdownTimeout: time.Second})
	}()
	<-started
	cancel()
	select {
	case err := <-errCh:
		if err == nil || err.Error() != "shutdown failed" {
			t.Fatalf("expected shutdown failed, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return")
	}
	close(released)
}

func TestRunShutdownThenListenError(t *testing.T) {
	restoreServerHooks(t)
	started := make(chan struct{})
	released := make(chan struct{})
	listenAndServe = func(*http.Server) error {
		close(started)
		<-released
		return errors.New("listen exploded")
	}
	shutdownServer = func(*http.Server, context.Context) error {
		close(released)
		return http.ErrServerClosed
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		a := &App{}
		a.Initialize()
		errCh <- a.Run(ctx, ServerConfig{Addr: "127.0.0.1:0", ShutdownTimeout: time.Second})
	}()
	<-started
	cancel()
	select {
	case err := <-errCh:
		if err == nil || err.Error() != "listen exploded" {
			t.Fatalf("expected listen exploded, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return")
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
