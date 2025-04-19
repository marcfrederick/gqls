package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
)

func main() {
	app := &cli.Command{
		Name:      "gqls",
		Usage:     "ConvertString a GQL schema into Neo4j format",
		ArgsUsage: "[input]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Sources:  cli.EnvVars("GQLS_INPUT"),
				Usage:    "Path to the input GQL schema file (use '-' to read from stdin)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "neo4j-version",
				Sources:  cli.EnvVars("GQLS_NEO4J_VERSION"),
				Usage:    "Specify the Neo4j version to use for conversion",
				Required: false,
			},
			&cli.BoolFlag{
				Name:        "apoc",
				Usage:       "Generate APOC triggers",
				DefaultText: "false",
			},
		},
		Action: run,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	inputPath := cmd.String("input")
	content, err := readInput(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	converted, err := gql2neo4j.ConvertBytes(content, gql2neo4j.Opts{
		Neo4jVersion: cmd.String("neo4j-version"),
		ApocEnabled:  cmd.Bool("apoc"),
	})
	if err != nil {
		return fmt.Errorf("failed to convert GQL to Neo4j format: %w", err)
	}

	fmt.Print(converted)
	return nil
}

func readInput(path string) ([]byte, error) {
	if path == "-" || path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}
