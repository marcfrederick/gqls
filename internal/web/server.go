// Package web provides a simple web server to serve the conversion tool
// and handle conversion requests.
package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
)

//go:embed static/*
var staticFiles embed.FS

// ConversionRequest represents the request body for the conversion API.
type ConversionRequest struct {
	GQL          string `json:"gql"`
	Neo4jVersion string `json:"neo4jVersion"`
	ApocEnabled  bool   `json:"apocEnabled"`
}

// ConversionResponse represents the response body for the conversion API.
type ConversionResponse struct {
	Neo4j string `json:"neo4j"`
	Error string `json:"error,omitempty"`
}

// Serve starts the web server on the specified address
func Serve(addr string) error {
	// Parse template
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		return err
	}

	// Handle static files
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle conversion API endpoint
	http.HandleFunc("/convert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ConversionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		neo4j, err := gql2neo4j.ConvertString(req.GQL, gql2neo4j.Opts{
			Neo4jVersion: req.Neo4jVersion,
			ApocEnabled:  req.ApocEnabled,
		})
		resp := ConversionResponse{
			Neo4j: neo4j,
		}
		if err != nil {
			resp.Error = err.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Handle static assets
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile(r.URL.Path[1:])
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		switch {
		case r.URL.Path[len(r.URL.Path)-4:] == ".css":
			w.Header().Set("Content-Type", "text/css")
		case r.URL.Path[len(r.URL.Path)-3:] == ".js":
			w.Header().Set("Content-Type", "application/javascript")
		}

		if _, err := w.Write(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting web server on %s", addr)
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return server.ListenAndServe()
}
