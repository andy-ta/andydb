package handler

import (
	"encoding/json"
	"fmt"
	"github.com/andy-ta/andydb/app/database"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	code := http.StatusInternalServerError
	var response interface{}
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]

	if result := database.Get(resourceName); result != nil {
		if response = result.Read(key); response != nil {
			respondJSON(w, http.StatusOK, response)
		} else {
			code = http.StatusNotFound
			err = fmt.Errorf("id %q does not exist", key)
		}
	} else {
		code = http.StatusNotFound
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, code, err.Error())
	}
}

func GetAll(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	code := http.StatusInternalServerError
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
		code = http.StatusNotFound
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, code, err.Error())
	}
}

func Update(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	code := http.StatusInternalServerError
	var response interface{}
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]

	if result := database.Get(resourceName); result != nil {
		body, err := getBody(r)
		if err == nil {
			if response = result.Update(key, body); response != nil {
				respondJSON(w, http.StatusOK, response)
			} else {
				code = http.StatusNotFound
				err = fmt.Errorf("id %q does not exist", key)
			}
		} else {
			code = http.StatusBadRequest
			err = fmt.Errorf("invalid body %q", body)
		}
	} else {
		code = http.StatusNotFound
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, code, err.Error())
	}
}

func Delete(w http.ResponseWriter, r *http.Request, database database.Resources) {
	var err error
	code := http.StatusInternalServerError
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]

	if result := database.Get(resourceName); result != nil {
		if deleted := result.Del(key); deleted {
			respondJSON(w, http.StatusNoContent, nil)
		} else {
			err = fmt.Errorf("failed to delete id %q", key)
		}
	} else {
		code = http.StatusBadRequest
		err = fmt.Errorf("resource %q does not exist", resourceName)
	}

	if err != nil {
		respondError(w, code, err.Error())
	}
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
	entry, err := getBody(r)
	result = database.Get(resourceName).Create(entry)

	if err == nil {
		respondJSON(w, http.StatusOK, result)
	} else {
		respondError(w, http.StatusConflict, err.Error())
	}
}

func getBody(r *http.Request) (interface{}, error) {
	var entry interface{}
	body, e := ioutil.ReadAll(r.Body)
	bodyString := string(body)
	e = json.Unmarshal([]byte(bodyString), &entry)
	return entry, e
}
