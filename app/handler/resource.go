package handler

import (
	"github.com/andy-ta/andydb/app/database"
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, "{ \"hello\": \"World\")")
}

func Create(w http.ResponseWriter, r *http.Request, database database.Resources) {
	respondJSON(w, http.StatusOK, "{ \"hello\": \"World\")")
}
