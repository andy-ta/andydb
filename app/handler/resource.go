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
	resource, ok := resolveResource(w, database, resourceName, true)
	if !ok {
		return
	}
	entry := resource.Read(key)
	if entry == nil {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("id %q does not exist", key))
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func GetAll(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	resource, ok := resolveResource(w, database, resourceName, true)
	if !ok {
		return
	}
	respondJSON(w, http.StatusOK, resource.ReadAll())
}

func Update(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]
	resource, ok := resolveResource(w, database, resourceName, true)
	if !ok {
		return
	}
	body, err := parseBody(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	entry, err := resource.Update(key, body)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func Delete(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	key := vars["id"]
	resource, ok := resolveResource(w, database, resourceName, true)
	if !ok {
		return
	}
	if deleted := resource.Del(key); !deleted {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("id %q does not exist", key))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Create(w http.ResponseWriter, r *http.Request, database *database.Resources) {
	vars := mux.Vars(r)
	resourceName := vars["resource"]
	resource, exists := resolveResource(w, database, resourceName, false)
	if !exists {
		if err := database.NewResource(resourceName); err != nil {
			resource = database.Get(resourceName)
			if resource == nil {
				RespondError(w, http.StatusConflict, err.Error())
				return
			}
		} else {
			resource = database.Get(resourceName)
		}
	}
	if resource == nil {
		RespondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to prepare resource %q", resourceName))
		return
	}
	body, err := parseBody(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := resource.Create(body)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, result)
}

func parseBody(r *http.Request) (map[string]interface{}, error) {
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
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	entry, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("request body must be a JSON object")
	}
	return entry, nil
}

func resolveResource(w http.ResponseWriter, database *database.Resources, resourceName string, respondIfMissing bool) (*database.Entries, bool) {
	resource := database.Get(resourceName)
	if resource == nil && respondIfMissing {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("resource %q does not exist", resourceName))
		return nil, false
	}
	return resource, resource != nil
}
