package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type store struct {
	mu    sync.Mutex
	notes []note
}

func newMux(s *store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, _ *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, http.StatusOK, append([]note{}, s.notes...))
	})
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil || in.Text == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "text is required"})
			return
		}
		s.mu.Lock()
		n := note{ID: len(s.notes) + 1, Text: in.Text}
		s.notes = append(s.notes, n)
		s.mu.Unlock()
		writeJSON(w, http.StatusCreated, n)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newMux(&store{}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
