package handler

import (
	"encoding/json"
	"fmt"
	"github.com/andy-ta/andydb/app/database"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, "{ \"hello\": \"World\")")
}

func Get(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	var response interface{}
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]

	if result := database.Get(resourceName); result != nil {
		if response = result.Read(key); response != nil {
			respondJSON(w, http.StatusOK, response)
		} else {
			err = fmt.Errorf("id %q does not exist", key)
		}
	} else {
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
	}
}

func GetAll(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	var response []interface{}
	vars := mux.Vars(r)
	resourceName := vars["resource"]

	if result := database.Get(resourceName); result != nil {
		if response = result.ReadAll(); response != nil {
			respondJSON(w, http.StatusOK, response)
		} else {
			err = fmt.Errorf("entries for %q not found", resourceName)
		}
	} else {
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
	}
}

// TODO: Error handling is so strange in Go?
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
	var entry interface{}
	body, e := ioutil.ReadAll(r.Body)
	err = e
	bodyString := string(body)
	e = json.Unmarshal([]byte(bodyString), &entry)
	result = database.Get(resourceName).Create(entry)

	if err == nil {
		respondJSON(w, http.StatusOK, result)
	} else {
		respondError(w, http.StatusConflict, err.Error())
	}
}
