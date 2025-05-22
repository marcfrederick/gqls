//go:build integration

package transmog

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var containerMutex sync.Mutex

func neo4jIntegrationTest(t *testing.T, testFunc func(session neo4j.SessionWithContext)) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	containerMutex.Lock()
	defer containerMutex.Unlock()

	container, err := startNeo4jContainer(t.Context())
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start Neo4j container: %v", err)
	}

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
	defer driver.Close(t.Context())

	if err := driver.VerifyConnectivity(t.Context()); err != nil {
		t.Fatalf("failed to verify connectivity: %v", err)
	}

	session := driver.NewSession(t.Context(), neo4j.SessionConfig{})
	defer session.Close(t.Context())

	testFunc(session)
}

func startNeo4jContainer(ctx context.Context) (testcontainers.Container, error) {
	var providerType testcontainers.ProviderType
	if os.Getenv("TESTCONTAINERS_DRIVER") == "podman" {
		providerType = testcontainers.ProviderPodman
	} else {
		providerType = testcontainers.ProviderDocker
	}

	req := testcontainers.GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "neo4j:2025.02-enterprise",
			ExposedPorts: []string{"7474/tcp", "7687/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithPort("7474/tcp").WithStartupTimeout(5 * time.Minute),
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
