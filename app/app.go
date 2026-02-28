package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/andy-ta/andydb/app/database"
	"github.com/andy-ta/andydb/app/handler"
	"github.com/gorilla/mux"
)

type App struct {
	Router   *mux.Router
	Database database.Resources
}

type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()
	a.Database = database.NewDatabase()
	a.setRouters()
}

func (a *App) setRouters() {
	r := a.Router.PathPrefix("/api").Subrouter()
	r.HandleFunc("/{resource}/{id}", a.handleRequest(handler.Get)).Methods("GET")
	r.HandleFunc("/{resource}", a.handleRequest(handler.GetAll)).Methods("GET")
	r.HandleFunc("/{resource}", a.handleRequest(handler.Create)).Methods("POST")
	r.HandleFunc("/{resource}/{id}", a.handleRequest(handler.Update)).Methods("PUT")
	r.HandleFunc("/{resource}/{id}", a.handleRequest(handler.Delete)).Methods("DELETE")
}

func (a *App) Run(ctx context.Context, cfg ServerConfig) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cfg.normalize()
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      a.Router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func (cfg *ServerConfig) normalize() {
	if cfg.Addr == "" {
		cfg.Addr = ":42069"
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 5 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 5 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 120 * time.Second
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}
}

type RequestHandlerFunction func(w http.ResponseWriter, r *http.Request, database database.Resources)

func (a *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, a.Database)
	}
}
