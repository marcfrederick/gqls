package gql2neo4j

import (
	"testing"
)

func Test_newNumericRangeTrigger(t *testing.T) {
	trigger, err := newNumericRangeTrigger(nodeEntityType, "Person", "age", 8)
	if err != nil {
		t.Fatalf("newNumericRangeTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("constraint_person_age_trigger", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND (e.age < -256 OR e.age > 255) CALL apoc.util.validate(true, 'age must be between -256 and 255', [])", {phase: "before"})`
	if got != want {
		t.Errorf("newNumericRangeTrigger() got = %s, want = %s", got, want)
	}
}

func Test_newStringLengthTrigger(t *testing.T) {
	trigger, err := newStringLengthTrigger(edgeEntityType, "Person", "name", 1, 255)
	if err != nil {
		t.Fatalf("newStringLengthTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("constraint_person_name_trigger", "UNWIND $createdRelationships AS e WITH e WHERE e:Person AND e.name IS NOT NULL AND (size(e.name) < 1 OR size(e.name) > 255) CALL apoc.util.validate(true, 'name must be between 1 and 255 characters', [])", {phase: "before"})`
	if got != want {
		t.Errorf("newStringLengthTrigger() got = %s, want = %s", got, want)
	}
}
