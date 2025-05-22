// Package main provides a command-line interface for converting GraphQL schemas to Neo4j format.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
	"github.com/marcfrederick/gqls/internal/gql2neo4j/transmog"
	"github.com/marcfrederick/gqls/internal/web"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	app := &cli.Command{
		Name:      "gqls",
		Usage:     "Convert a GQL schema into Neo4j format",
		ArgsUsage: "[input]",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Sources: cli.EnvVars("GQLS_INPUT"),
				Usage:   "Path to the input GQL schema file (use '-' to read from stdin)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Sources: cli.EnvVars("GQLS_OUTPUT"),
				Usage:   "Path to the output file (defaults to stdout)",
			},
			&cli.StringFlag{
				Name:        "neo4j-version",
				Sources:     cli.EnvVars("GQLS_NEO4J_VERSION"),
				Usage:       "Specify the Neo4j version to use for conversion",
				DefaultText: "latest",
			},
			&cli.BoolFlag{
				Name:        "apoc",
				Usage:       "Generate APOC triggers",
				DefaultText: "false",
			},
			// BEGIN AI SECTION
			// @ai-generated model="claude-3-7-sonnet-20250219"
			&cli.StringFlag{
				Name:  "serve",
				Usage: "Start web server on the specified address (e.g., :8080)",
			},
			// END AI SECTION
		},
		Action: run,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	// BEGIN AI SECTION
	// @ai-generated model="claude-3-7-sonnet-20250219"
	if addr := cmd.String("serve"); addr != "" {
		return web.Serve(addr)
	}
	// END AI SECTION

	inputPath := cmd.String("input")
	content, err := readInput(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	converted, err := gql2neo4j.ConvertBytes(content, transmog.DatabaseMetadata{
		Version:     cmd.String("neo4j-version"),
		APOCEnabled: cmd.Bool("apoc"),
	})
	if err != nil {
		return fmt.Errorf("failed to convert GQL to Neo4j format: %w", err)
	}

	if outputPath := cmd.String("output"); outputPath != "" {
		if err := os.WriteFile(outputPath, converted, 0o644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(converted); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}

func readInput(path string) ([]byte, error) {
	if path == "-" || path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(filepath.Clean(path))
}
