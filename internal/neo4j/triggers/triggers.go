package triggers

import (
	"fmt"
	"strconv"

	"github.com/marcfrederick/gqls/internal/neo4j"
)

// Phase represents the phase of a trigger in Neo4j.
type Phase string

const (
	Before   Phase = "before"
	After    Phase = "after"
	Rollback Phase = "rollback"
	Async    Phase = "async"
)

// Selector represents the selector for a Neo4j APOC trigger.
type Selector struct {
	// Phase is the phase in which the trigger is executed.
	Phase Phase `json:"phase"`
}

// Trigger represents a Neo4j APOC trigger.
type Trigger struct {
	Name              string
	KernelTransaction string
	Selector          Selector
}

var _ neo4j.SchemaObject = (*Trigger)(nil)

func (t *Trigger) CreateStatement() string {
	// FIXME: Should probably use apoc.trigger.install instead of apoc.trigger.add
	return fmt.Sprintf(
		"CALL apoc.trigger.add(%s, %s, %s)",
		strconv.Quote(t.Name),
		strconv.Quote(t.KernelTransaction),
		fmt.Sprintf("{phase: %s}", strconv.Quote(string(t.Selector.Phase))),
	)
}
