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
	want := `CALL apoc.trigger.add("numeric_range_person_age", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND (e.age < -256 OR e.age > 255) CALL apoc.util.validate(true, 'age must be between -256 and 255', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.age IS NOT NULL AND (change.age < -256 OR change.age > 255) CALL apoc.util.validate(true, 'age must be between -256 and 255', []) RETURN null", {phase: "before"})`
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
	want := `CALL apoc.trigger.add("string_length_person_name", "UNWIND $createdRelationships AS e WITH e WHERE e:Person AND e.name IS NOT NULL AND (size(e.name) < 1 OR size(e.name) > 255) CALL apoc.util.validate(true, 'name must be between 1 and 255 characters', []) RETURN null UNION UNWIND keys($assignedRelationshipProperties) AS prop UNWIND $assignedRelationshipProperties[prop] AS change WITH change WHERE change:Person AND change.name IS NOT NULL AND (size(change.name) < 1 OR size(change.name) > 255) CALL apoc.util.validate(true, 'name must be between 1 and 255 characters', []) RETURN null", {phase: "before"})`
	if got != want {
		t.Errorf("newStringLengthTrigger() got = %s, want = %s", got, want)
	}
}

func Test_newEdgeEndpointTrigger(t *testing.T) {
	trigger, err := newEdgeEndpointTrigger("LIKES", "Person", "Post")
	if err != nil {
		t.Fatalf("newEdgeEndpointTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("edge_endpoint_likes_person_post", "UNWIND $createdRelationships AS rel WITH rel WHERE type(rel) = \"LIKES\" AND NOT (startNode(rel):Person AND endNode(rel):Post) CALL apoc.util.validate( true, \"LIKES relationship must be between Person and Post\", [] ) RETURN null", {phase: "before"})`
	if got != want {
		t.Errorf("newEdgeEndpointTrigger() got = %s, want = %s", got, want)
	}
}
