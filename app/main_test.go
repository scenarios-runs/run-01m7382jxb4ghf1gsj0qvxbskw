package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(&store{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCreateAndList(t *testing.T) {
	mux := newMux(&store{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"text":"buy milk"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", rec.Code)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes", nil))
	if !strings.Contains(rec.Body.String(), "buy milk") {
		t.Fatalf("list body = %q, want the note", rec.Body.String())
	}
}

func TestRejectsEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(&store{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"text":""}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}
