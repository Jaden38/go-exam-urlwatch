package api_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jaden38/go-exam-urlwatch/internal/api"
	"github.com/Jaden38/go-exam-urlwatch/internal/checker"
	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
	"github.com/Jaden38/go-exam-urlwatch/internal/store"
)

func newTestServer() http.Handler {
	m := &checker.Mock{Results: map[string]domain.CheckResult{
		"https://go.dev": {StatusCode: 200, OK: true, LatencyMs: 12},
	}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return api.NewServer(m, store.NewMemory(), logger).Routes()
}

func TestCreateBatchSuccess(t *testing.T) {
	srv := newTestServer()
	body := `{"urls":["https://go.dev"],"options":{"concurrency":2,"timeout_ms":1000}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/checks", strings.NewReader(body))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201 ; corps = %s", rec.Code, rec.Body.String())
	}
	var batch domain.Batch
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatal(err)
	}
	if batch.ID == "" {
		t.Error("batch_id vide")
	}
	if batch.Summary.Total != 1 || batch.Summary.Up != 1 {
		t.Errorf("résumé inattendu : %+v", batch.Summary)
	}
}

func TestGetBatchRoundTrip(t *testing.T) {
	srv := newTestServer()

	createReq := httptest.NewRequest(http.MethodPost, "/v1/checks", strings.NewReader(`{"urls":["https://go.dev"]}`))
	createRec := httptest.NewRecorder()
	srv.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("création : code = %d", createRec.Code)
	}
	var created domain.Batch
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/checks/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	srv.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("lecture : code = %d", getRec.Code)
	}
	var got domain.Batch
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID {
		t.Errorf("id relu = %q, want %q", got.ID, created.ID)
	}
}

func TestGetBatchNotFound(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/checks/b_inconnu", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
	var eb struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &eb); err != nil {
		t.Fatal(err)
	}
	if eb.Error.Code != "batch_not_found" {
		t.Errorf("code d'erreur = %q, want batch_not_found", eb.Error.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodDelete, "/v1/checks/b_x", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("code = %d, want 405", rec.Code)
	}
	var eb struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &eb); err != nil {
		t.Fatalf("le corps 405 doit être du JSON : %v", err)
	}
	if eb.Error.Code != "method_not_allowed" {
		t.Errorf("code d'erreur = %q, want method_not_allowed", eb.Error.Code)
	}
}

func TestCreateBatchValidation(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/checks", strings.NewReader(`{"urls":[]}`))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}
