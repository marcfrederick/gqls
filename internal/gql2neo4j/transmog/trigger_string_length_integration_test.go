//go:build integration

package transmog

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestStringLengthTrigger(t *testing.T) {
	neo4jIntegrationTest(t, func(session neo4j.SessionWithContext) {
		trigger, err := newStringLengthTrigger(EdgeEntityType, "Person", "name", 1, 255)
		if err != nil {
			t.Fatalf("failed to create string length trigger: %v", err)
		}

		_, err = session.Run(t.Context(), trigger.CreateStatement(), nil)
		if err != nil {
			t.Errorf("failed to run trigger creation statement: %v", err)
		}
	})
}
