package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/andy-ta/andydb/app/database"
	"github.com/gorilla/mux"
)

const maxRequestBodySize = 1 << 20

func Get(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]
	resource := database.Get(resourceName)
	if resource == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("resource %q does not exist", resourceName))
		return
	}
	entry := resource.Read(key)
	if entry == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("id %q does not exist", key))
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func GetAll(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	resource := database.Get(resourceName)
	if resource == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("resource %q does not exist", resourceName))
		return
	}
	respondJSON(w, http.StatusOK, resource.ReadAll())
}

func Update(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]
	resource := database.Get(resourceName)
	if resource == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("resource %q does not exist", resourceName))
		return
	}
	body, err := parseBody(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	entry := resource.Update(key, body)
	if entry == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("id %q does not exist", key))
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func Delete(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]
	resource := database.Get(resourceName)
	if resource == nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("resource %q does not exist", resourceName))
		return
	}
	if deleted := resource.Del(key); !deleted {
		respondError(w, http.StatusNotFound, fmt.Sprintf("id %q does not exist", key))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Create(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	resource := database.Get(resourceName)
	if resource == nil {
		if err := database.NewResource(resourceName); err != nil {
			resource = database.Get(resourceName)
			if resource == nil {
				respondError(w, http.StatusConflict, err.Error())
				return
			}
		} else {
			resource = database.Get(resourceName)
		}
	}
	if resource == nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to prepare resource %q", resourceName))
		return
	}
	body, err := parseBody(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	result := resource.Create(body)
	respondJSON(w, http.StatusCreated, result)
}

func parseBody(r *http.Request) (interface{}, error) {
	limitedReader := io.LimitReader(r.Body, maxRequestBodySize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}
	if len(body) > maxRequestBodySize {
		return nil, fmt.Errorf("request body exceeds %d bytes", maxRequestBodySize)
	}
	var entry interface{}
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return entry, nil
}
