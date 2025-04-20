package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
)

//go:embed static/*
var staticFiles embed.FS

type ConversionRequest struct {
	GQL          string `json:"gql"`
	Neo4jVersion string `json:"neo4jVersion"`
	ApocEnabled  bool   `json:"apocEnabled"`
}

type ConversionResponse struct {
	Neo4j string `json:"neo4j"`
	Error string `json:"error,omitempty"`
}

type Neo4jVersionsResponse struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Versions map[string]struct {
		Dist struct {
			Mac   string `json:"mac"`
			Win   string `json:"win"`
			Linux string `json:"linux"`
		} `json:"dist"`
	} `json:"versions"`
}

// Serve starts the web server on the specified address
func Serve(addr string) error {
	// Parse template
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		return err
	}

	// Handle static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle Neo4j versions proxy endpoint
	http.HandleFunc("/api/neo4j-versions", func(w http.ResponseWriter, r *http.Request) {
		// Fetch versions from Neo4j
		resp, err := http.Get("https://dist.neo4j.org/versions/v2/neo4j-versions.json")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch versions: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Parse the response
		var versionsResp Neo4jVersionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&versionsResp); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse versions: %v", err), http.StatusInternalServerError)
			return
		}

		// Get all versions and filter to major.minor
		uniqueVersions := make(map[string]struct{})
		for version := range versionsResp.Versions {
			// Extract major.minor for both calendar and semver versions
			if match := regexp.MustCompile(`^(\d{4}\.\d{2}|\d+\.\d+)`).FindString(version); match != "" {
				uniqueVersions[match] = struct{}{}
			}
		}

		// Convert to slice
		versions := make([]string, 0, len(uniqueVersions))
		for version := range uniqueVersions {
			versions = append(versions, version)
		}

		// Sort versions in descending order
		sort.Slice(versions, func(i, j int) bool {
			// Split versions into components
			partsI := strings.Split(versions[i], ".")
			partsJ := strings.Split(versions[j], ".")

			// Compare each component
			for k := 0; k < len(partsI) && k < len(partsJ); k++ {
				var numI, numJ int
				fmt.Sscanf(partsI[k], "%d", &numI)
				fmt.Sscanf(partsJ[k], "%d", &numJ)
				if numI != numJ {
					return numI > numJ
				}
			}
			// If all components match, longer version is considered greater
			return len(partsI) > len(partsJ)
		})

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(versions); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
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
		w.Write(data)
	})

	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
