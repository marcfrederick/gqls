//go:build integration

package convert

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestEdgeEndpointTrigger(t *testing.T) {
	neo4jIntegrationTest(t, func(session neo4j.SessionWithContext) {
		trigger, err := newEdgeEndpointTrigger("LIKES", "Person", "Post")
		if err != nil {
			t.Fatalf("failed to create edge endpoint trigger: %v", err)
		}

		_, err = session.Run(t.Context(), trigger.CreateStatement(), nil)
		if err != nil {
			t.Errorf("failed to run trigger creation statement: %v", err)
		}
	})
}
