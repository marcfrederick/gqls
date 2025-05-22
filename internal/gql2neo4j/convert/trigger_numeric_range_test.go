package convert

import (
	"testing"
)

func Test_newNumericRangeTrigger(t *testing.T) {
	trigger, err := newNumericRangeTrigger(NodeEntityType, "Person", "age", 8)
	if err != nil {
		t.Fatalf("newNumericRangeTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("numeric_range_person_age", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND (e.age < -256 OR e.age > 255) CALL apoc.util.validate(true, 'age must be between -256 and 255', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.age IS NOT NULL AND (change.age < -256 OR change.age > 255) CALL apoc.util.validate(true, 'age must be between -256 and 255', []) RETURN null", {phase: "before"})`
	if got != want {
		t.Errorf("newNumericRangeTrigger() got = %s, want = %s", got, want)
	}
}
