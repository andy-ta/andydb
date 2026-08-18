package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/andy-ta/andydb/app/database"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
)

func TestParseBodyValidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"andy"}`))
	body, err := parseBody(req)
	if err != nil {
		t.Fatalf("parseBody returned error: %v", err)
	}
	if body["name"] != "andy" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestParseBodyTooLarge(t *testing.T) {
	large := strings.Repeat("a", maxRequestBodySize+1)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(large))
	if _, err := parseBody(req); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestParseBodyInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{notjson}"))
	if _, err := parseBody(req); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestParseBodyEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	if _, err := parseBody(req); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty body error, got %v", err)
	}
}

func TestParseBodyRequiresJSONObject(t *testing.T) {
	cases := []string{`["not","an","object"]`, `"string"`, `42`, `true`, `null`}
	for _, payload := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
		if _, err := parseBody(req); err == nil || !strings.Contains(err.Error(), "JSON object") {
			t.Fatalf("payload %s: expected JSON object error, got %v", payload, err)
		}
	}
}

func TestResolveResourceNotFound(t *testing.T) {
	db := database.NewDatabase()
	recorder := httptest.NewRecorder()
	res, ok := resolveResource(recorder, db, "missing", true)
	if ok || res != nil {
		t.Fatalf("expected missing result, got %v", res)
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestResolveResourceFound(t *testing.T) {
	db := database.NewDatabase()
	if err := db.NewResource("found"); err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}
	recorder := httptest.NewRecorder()
	res, ok := resolveResource(recorder, db, "found", true)
	if !ok || res == nil {
		t.Fatalf("expected resource, got %v", res)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected no response body, got %q", recorder.Body.String())
	}
}

func TestCRUDHandlers(t *testing.T) {
	db := database.NewDatabase()
	resourceName := "contacts"

	first := createEntry(t, db, resourceName, `{"name":"andy"}`)
	second := createEntry(t, db, resourceName, `{"name":"betty"}`)

	if first["_id"] == second["_id"] {
		t.Fatalf("expected distinct ids, got %q", first["_id"])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/contacts/"+first["_id"].(string), nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"resource": resourceName, "id": first["_id"].(string)})
	getRec := httptest.NewRecorder()
	Get(getRec, getReq, db)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d (%s)", getRec.Code, getRec.Body.String())
	}
	if ct := getRec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	got := decodeMap(t, getRec)
	if got["name"] != "andy" {
		t.Fatalf("expected name=andy, got %v", got["name"])
	}
	if got["_id"] != first["_id"] {
		t.Fatalf("expected same id on get, got %v", got["_id"])
	}

	getAllReq := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	getAllReq = mux.SetURLVars(getAllReq, map[string]string{"resource": resourceName})
	getAllRec := httptest.NewRecorder()
	GetAll(getAllRec, getAllReq, db)
	if getAllRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get all, got %d (%s)", getAllRec.Code, getAllRec.Body.String())
	}
	list := decodeSlice(t, getAllRec)
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	if !hasID(list, first["_id"]) || !hasID(list, second["_id"]) {
		t.Fatalf("expected get all to include both created ids, got %#v", list)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/contacts/"+first["_id"].(string), strings.NewReader(`{"name":"andy-updated"}`))
	updateReq = mux.SetURLVars(updateReq, map[string]string{"resource": resourceName, "id": first["_id"].(string)})
	updateRec := httptest.NewRecorder()
	Update(updateRec, updateReq, db)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", updateRec.Code)
	}
	updated := decodeMap(t, updateRec)
	if updated["_id"] != first["_id"] {
		t.Fatalf("expected id unchanged, got %v", updated["_id"])
	}
	if updated["name"] != "andy-updated" {
		t.Fatalf("expected updated name, got %v", updated["name"])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/contacts/"+first["_id"].(string), nil)
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"resource": resourceName, "id": first["_id"].(string)})
	deleteRec := httptest.NewRecorder()
	Delete(deleteRec, deleteReq, db)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", deleteRec.Code)
	}

	confirmReq := httptest.NewRequest(http.MethodGet, "/api/contacts/"+first["_id"].(string), nil)
	confirmReq = mux.SetURLVars(confirmReq, map[string]string{"resource": resourceName, "id": first["_id"].(string)})
	confirmRec := httptest.NewRecorder()
	Get(confirmRec, confirmReq, db)
	if confirmRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", confirmRec.Code)
	}
	confirmErr := decodeError(t, confirmRec)
	if !strings.Contains(confirmErr, first["_id"].(string)) {
		t.Fatalf("expected not-found message to include missing id, got %q", confirmErr)
	}
}

func TestCreateNonObjectBodyReturnsBadRequest(t *testing.T) {
	db := database.NewDatabase()
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", strings.NewReader(`["not","an","object"]`))
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts"})
	rec := httptest.NewRecorder()

	Create(rec, req, db)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-object create body, got %d", rec.Code)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "JSON object") {
		t.Fatalf("expected JSON object error, got %q", got)
	}
}

func TestUpdateNonObjectBodyReturnsBadRequest(t *testing.T) {
	db := database.NewDatabase()
	created := createEntry(t, db, "contacts", `{"name":"andy"}`)
	id := created["_id"].(string)

	req := httptest.NewRequest(http.MethodPut, "/api/contacts/"+id, strings.NewReader(`42`))
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts", "id": id})
	rec := httptest.NewRecorder()
	Update(rec, req, db)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-object update body, got %d", rec.Code)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "JSON object") {
		t.Fatalf("expected JSON object error, got %q", got)
	}
}

func TestDeleteMissingIDReturnsNotFound(t *testing.T) {
	db := database.NewDatabase()
	if err := db.NewResource("contacts"); err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/contacts/missing-id", nil)
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts", "id": "missing-id"})
	rec := httptest.NewRecorder()
	Delete(rec, req, db)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing id, got %d", rec.Code)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "missing-id") {
		t.Fatalf("expected not-found message to include missing id, got %q", got)
	}
}

func TestCreateInvalidBodyReturnsBadRequest(t *testing.T) {
	db := database.NewDatabase()
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", strings.NewReader("{invalid-json"))
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts"})
	rec := httptest.NewRecorder()

	Create(rec, req, db)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid create body, got %d", rec.Code)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "invalid JSON") {
		t.Fatalf("expected invalid JSON error, got %q", got)
	}
}

func TestCreateIgnoresClientID(t *testing.T) {
	db := database.NewDatabase()
	created := createEntry(t, db, "contacts", `{"name":"andy","_id":"client-chosen"}`)
	id, ok := created["_id"].(string)
	if !ok || id == "" || id == "client-chosen" {
		t.Fatalf("expected server-assigned _id, got %#v", created["_id"])
	}
	if _, err := uuid.FromString(id); err != nil {
		t.Fatalf("expected UUID _id, got %q: %v", id, err)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/api/contacts/client-chosen", nil)
	missingReq = mux.SetURLVars(missingReq, map[string]string{"resource": "contacts", "id": "client-chosen"})
	missingRec := httptest.NewRecorder()
	Get(missingRec, missingReq, db)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for client-chosen id, got %d", missingRec.Code)
	}

	gotReq := httptest.NewRequest(http.MethodGet, "/api/contacts/"+id, nil)
	gotReq = mux.SetURLVars(gotReq, map[string]string{"resource": "contacts", "id": id})
	gotRec := httptest.NewRecorder()
	Get(gotRec, gotReq, db)
	if gotRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for minted id, got %d", gotRec.Code)
	}
}

func TestUpdateIgnoresClientID(t *testing.T) {
	db := database.NewDatabase()
	created := createEntry(t, db, "contacts", `{"name":"andy"}`)
	id := created["_id"].(string)

	req := httptest.NewRequest(http.MethodPut, "/api/contacts/"+id, strings.NewReader(`{"name":"betty","_id":"other"}`))
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts", "id": id})
	rec := httptest.NewRecorder()
	Update(rec, req, db)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d (%s)", rec.Code, rec.Body.String())
	}
	updated := decodeMap(t, rec)
	if updated["_id"] != id {
		t.Fatalf("expected server _id to stay %q, got %#v", id, updated["_id"])
	}

	otherReq := httptest.NewRequest(http.MethodGet, "/api/contacts/other", nil)
	otherReq = mux.SetURLVars(otherReq, map[string]string{"resource": "contacts", "id": "other"})
	otherRec := httptest.NewRecorder()
	Get(otherRec, otherReq, db)
	if otherRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for client _id, got %d", otherRec.Code)
	}
}

func TestUpdateMissingIDReturnsNotFound(t *testing.T) {
	db := database.NewDatabase()
	if err := db.NewResource("contacts"); err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/contacts/missing-id", strings.NewReader(`{"name":"andy"}`))
	req = mux.SetURLVars(req, map[string]string{"resource": "contacts", "id": "missing-id"})
	rec := httptest.NewRecorder()
	Update(rec, req, db)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing id, got %d", rec.Code)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "missing-id") {
		t.Fatalf("expected not-found message to include missing id, got %q", got)
	}
	if got := len(db.Get("contacts").ReadAll()); got != 0 {
		t.Fatalf("PUT of a missing id must not create a row, got %d", got)
	}

	confirmReq := httptest.NewRequest(http.MethodGet, "/api/contacts/missing-id", nil)
	confirmReq = mux.SetURLVars(confirmReq, map[string]string{"resource": "contacts", "id": "missing-id"})
	confirmRec := httptest.NewRecorder()
	Get(confirmRec, confirmReq, db)
	if confirmRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on follow-up GET, got %d", confirmRec.Code)
	}
}

func TestConcurrentCreateSameResource(t *testing.T) {
	db := database.NewDatabase()
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	codes := make([]int, workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/contacts", strings.NewReader(fmt.Sprintf(`{"n":%d}`, i)))
			req = mux.SetURLVars(req, map[string]string{"resource": "contacts"})
			rec := httptest.NewRecorder()
			Create(rec, req, db)
			codes[i] = rec.Code
		}(i)
	}
	wg.Wait()

	for i, code := range codes {
		if code != http.StatusCreated {
			t.Fatalf("worker %d: expected 201, got %d", i, code)
		}
	}
	resource := db.Get("contacts")
	if resource == nil {
		t.Fatal("expected contacts resource")
	}
	if got := len(resource.ReadAll()); got != workers {
		t.Fatalf("expected %d entries, got %d", workers, got)
	}
}

func TestRespondJSONMarshalErrorIsJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	respondJSON(rec, http.StatusOK, make(chan int))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "failed to encode") {
		t.Fatalf("unexpected error payload: %q", got)
	}
}

func TestGetMissingResourceReturnsNotFoundErrorPayload(t *testing.T) {
	db := database.NewDatabase()
	req := httptest.NewRequest(http.MethodGet, "/api/missing/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"resource": "missing", "id": "abc"})
	rec := httptest.NewRecorder()

	Get(rec, req, db)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing resource, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	if got := decodeError(t, rec); !strings.Contains(got, "resource \"missing\" does not exist") {
		t.Fatalf("unexpected error payload: %q", got)
	}
}

func createEntry(t *testing.T, db *database.Resources, resourceName, payload string) map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/"+resourceName, strings.NewReader(payload))
	req = mux.SetURLVars(req, map[string]string{"resource": resourceName})
	rec := httptest.NewRecorder()
	Create(rec, req, db)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create, got %d", rec.Code)
	}
	return decodeMap(t, rec)
}

func decodeMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var value map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &value); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return value
}

func decodeSlice(t *testing.T, rec *httptest.ResponseRecorder) []map[string]interface{} {
	t.Helper()
	var values []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &values); err != nil {
		t.Fatalf("failed to decode response slice: %v", err)
	}
	return values
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	decoded := decodeMap(t, rec)
	err, ok := decoded["error"].(string)
	if !ok {
		t.Fatalf("expected error string field, got %#v", decoded)
	}
	return err
}

func hasID(values []map[string]interface{}, id interface{}) bool {
	for _, value := range values {
		if value["_id"] == id {
			return true
		}
	}
	return false
}
