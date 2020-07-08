package handler

import "net/http"

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, "{ \"hello\": \"World\")")
}
