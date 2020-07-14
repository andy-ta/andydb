package handler

import (
	"encoding/json"
	"github.com/andy-ta/andydb/app/database"
	"github.com/bennyscetbun/jsongo"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, "{ \"hello\": \"World\")")
}

func Create(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	var result interface{}
	vars := mux.Vars(r)
	resourceName := vars["resource"]

	// If the resource does not exist yet.
	if !database.Exists(resourceName) {
		r, e := database.NewResource(vars["resource"])
		result = r
		err = e
	}

	// Create the entry
	root := jsongo.Node{}
	body, e := ioutil.ReadAll(r.Body)
	err = e
	bodyString := string(body)
	e = json.Unmarshal([]byte(bodyString), &root)
	result = database.Get(resourceName).Create(root)

	if err == nil {
		respondJSON(w, http.StatusOK, result)
	} else {
		respondError(w, http.StatusConflict, err.Error())
	}
}
