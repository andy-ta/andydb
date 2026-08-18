package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/andy-ta/andydb/app/database"
	"github.com/andy-ta/andydb/app/handler"
	"github.com/gorilla/mux"
)

type App struct {
	Router   *mux.Router
	Database *database.Resources
}

type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	DevMode         bool
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
	mode := "production"
	if cfg.DevMode {
		mode = "development"
	}
	log.Printf("AndyDB running on %s (%s mode)", cfg.Addr, mode)
	rootHandler := http.Handler(a.Router)
	rootHandler = a.recoverer(rootHandler)
	if cfg.DevMode {
		rootHandler = a.requestLogger(rootHandler)
	}
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rootHandler,
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

type RequestHandlerFunction func(w http.ResponseWriter, r *http.Request, database *database.Resources)

func (a *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, a.Database)
	}
}

func (a *App) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEV] %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (a *App) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				handler.RespondError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
