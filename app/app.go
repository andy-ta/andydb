package app

import (
	"log"
	"net/http"

	"github.com/andy-ta/andydb/app/database"
	"github.com/andy-ta/andydb/app/handler"
	"github.com/gorilla/mux"
)

type App struct {
	Router   *mux.Router
	Database database.Resources
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
	r.HandleFunc("/{resource}/{id}", handler.HelloWorld).Methods("PUT")
	r.HandleFunc("/{resource}/{id}", handler.HelloWorld).Methods("DELETE")
}

func (a *App) Run(host string) {
	log.Fatal(http.ListenAndServe(host, a.Router))
}

type RequestHandlerFunction func(w http.ResponseWriter, r *http.Request, database database.Resources)

func (a *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, a.Database)
	}
}
