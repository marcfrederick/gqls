//go:build wasm
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
	"github.com/marcfrederick/gqls/internal/gql2neo4j/convert"
)

var (
	version = "dev"
	commit  = "none"
)

// ConversionRequest represents the request for the conversion function.
type ConversionRequest struct {
	GQL          string `json:"gql"`
	Neo4jVersion string `json:"neo4jVersion"`
	ApocEnabled  bool   `json:"apocEnabled"`
}

// ConversionResponse represents the response from the conversion function.
type ConversionResponse struct {
	Neo4j string `json:"neo4j"`
	Error string `json:"error,omitempty"`
}

func main() {
	js.Global().Set("convertGQL", js.FuncOf(convertGQL))
	js.Global().Set("getVersionInfo", js.FuncOf(getVersionInfo))
	select {}
}

func convertGQL(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return map[string]interface{}{
			"error": "Invalid number of arguments, expected 1",
		}
	}

	// Parse the JSON input
	jsonStr := args[0].String()
	var req ConversionRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return map[string]interface{}{
			"error": "Failed to parse request: " + err.Error(),
		}
	}

	// Perform the conversion
	neo4j, err := gql2neo4j.ConvertString(req.GQL, convert.DatabaseMetadata{
		Version:     req.Neo4jVersion,
		APOCEnabled: req.ApocEnabled,
	})

	resp := ConversionResponse{
		Neo4j: neo4j,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	// Convert response back to JavaScript object
	respJSON, _ := json.Marshal(resp)
	return js.Global().Get("JSON").Call("parse", string(respJSON))
}

func getVersionInfo(this js.Value, args []js.Value) interface{} {
	return version
}
