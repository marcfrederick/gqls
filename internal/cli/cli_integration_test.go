//go:build integration

package cli_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/marcfrederick/gqls/internal/cli"
)

// neo4jVersion is the version of Neo4j to use in the integration tests.
const neo4jVersion = "2025.04-enterprise"

var containerMutex sync.Mutex

type testCase struct {
	Name    string `yaml:"name"`
	Query   string `yaml:"query"`
	WantErr bool   `yaml:"want_err"`
}

type testSpec struct {
	Schema string     `yaml:"schema"`
	Cases  []testCase `yaml:"cases"`
}

func TestGQLSchemaConversion(t *testing.T) {
	testSpecPaths, err := filepath.Glob("testdata/*.spec.yml")
	if err != nil {
		t.Fatalf("failed to find test cases: %v", err)
	}

	for _, testSpecPath := range testSpecPaths {
		name := filepath.Base(testSpecPath)
		spec := loadTestSpec(t, testSpecPath)

		neo4jIntegrationTest(t, func(session neo4j.SessionWithContext) {
			// Create the schema in Neo4j
			statements := convertGQLToNeo4j(t, spec.Schema)
			for _, statement := range splitStatements(statements) {
				t.Logf("Creating schema statement for spec %q: %s", name, statement)
				if _, err := session.Run(t.Context(), statement, nil); err != nil {
					t.Fatalf("failed to create statement for spec %q: %v", name, err)
				}
			}

			// Run test cases
			t.Run(name, func(t *testing.T) {
				for _, tCase := range spec.Cases {
					t.Run(tCase.Name, func(t *testing.T) {
						var anyError error
						for _, statement := range splitStatements(tCase.Query) {
							res, err := session.Run(t.Context(), statement, nil)
							if err == nil {
								// Consume the result to check for APOC trigger errors.
								_, err = res.Consume(t.Context())
							}
							if err != nil && anyError == nil {
								anyError = err
							}
						}

						if tCase.WantErr {
							if anyError == nil {
								t.Error("expected error, but got none")
							}
						} else if anyError != nil {
							t.Errorf("unexpected error: %v", anyError)
						}
					})
				}
			})
		})
	}
}

// loadTestSpec loads a test specification from a YAML file.
func loadTestSpec(t *testing.T, path string) testSpec {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read test file %q: %v", path, err)
	}

	var tc testSpec
	if err := yaml.Unmarshal(data, &tc); err != nil {
		t.Fatalf("failed to unmarshal test file %q: %v", path, err)
	}
	return tc
}

// splitStatements splits a string into individual statements based on semicolons.
func splitStatements(s string) []string {
	var statements []string
	for _, statement := range strings.Split(s, ";") {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}

// convertGQLToNeo4j uses the gqls CLI to convert a GQL schema to Neo4j Cypher statements.
func convertGQLToNeo4j(t *testing.T, schema string) string {
	t.Helper()

	inputPath := path.Join(t.TempDir(), "schema.gql")
	if err := os.WriteFile(inputPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("failed to write schema to %q: %v", inputPath, err)
	}

	outputPath := path.Join(t.TempDir(), "converted.cypher")

	cli.Run(t.Context(), "<version>", "<commit>", []string{
		"gqls",
		"--input", inputPath,
		"--output", outputPath,
		"--neo4j-version", "latest",
		"--apoc",
	})

	converted, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read converted file %q: %v", outputPath, err)
	}
	return string(converted)
}

func neo4jIntegrationTest(t *testing.T, testFunc func(session neo4j.SessionWithContext)) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	containerMutex.Lock()
	defer containerMutex.Unlock()

	container, err := startNeo4jContainer(t.Context())
	if err != nil {
		t.Fatalf("failed to start Neo4j container: %v", err)
	}
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(t.Context())
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(t.Context(), "7687")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	driver, err := neo4j.NewDriverWithContext(fmt.Sprintf("bolt://%s:%d", host, port.Int()), neo4j.NoAuth())
	if err != nil {
		t.Fatalf("failed to create Neo4j driver: %v", err)
	}
	defer func() {
		if closeErr := driver.Close(t.Context()); closeErr != nil {
			t.Errorf("failed to close Neo4j driver: %v", closeErr)
		}
	}()

	if err := driver.VerifyConnectivity(t.Context()); err != nil {
		t.Fatalf("failed to verify connectivity: %v", err)
	}

	session := driver.NewSession(t.Context(), neo4j.SessionConfig{})
	defer func() {
		if closeErr := session.Close(t.Context()); closeErr != nil {
			t.Errorf("failed to close Neo4j session: %v", closeErr)
		}
	}()

	testFunc(session)
}

// startNeo4jContainer starts a Neo4j container with the specified configuration.
func startNeo4jContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		ProviderType: getProviderType(),
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.io/neo4j:" + neo4jVersion,
			ExposedPorts: []string{"7687/tcp"},
			WaitingFor:   wait.ForLog("Bolt enabled on"),
			Env: map[string]string{
				"NEO4J_AUTH":                       "none",
				"NEO4J_ACCEPT_LICENSE_AGREEMENT":   "yes",
				"NEO4J_dbms_usage__report_enabled": "false",
				"NEO4JLABS_PLUGINS":                "[\"apoc\"]",
				"NEO4J_apoc_trigger_enabled":       "true",
			},
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return container, nil
}

// getProviderType returns the provider type based on the TESTCONTAINERS_DRIVER environment variable.
func getProviderType() testcontainers.ProviderType {
	if os.Getenv("TESTCONTAINERS_DRIVER") == "podman" {
		return testcontainers.ProviderPodman
	}
	return testcontainers.ProviderDocker
}
