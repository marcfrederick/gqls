//go:build integration

package convert

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestCheckConstraintTrigger(t *testing.T) {
	neo4jIntegrationTest(t, func(session neo4j.SessionWithContext) {
		trigger, err := newCheckConstraintTrigger(NodeEntityType, "Person", "age", "age >= 0 AND age <= 120")
		if err != nil {
			t.Fatalf("failed to create check constraint trigger: %v", err)
		}

		_, err = session.Run(t.Context(), trigger.CreateStatement(), nil)
		if err != nil {
			t.Errorf("failed to run trigger creation statement: %v", err)
		}
	})
}
