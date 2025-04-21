// Package triggers provides functionality for creating and managing triggers in Neo4j.
package triggers

import (
	"fmt"
	"strconv"

	"github.com/marcfrederick/gqls/internal/neo4j"
)

// Phase represents the phase of a trigger in Neo4j.
type Phase string

const (
	// Before represents the phase before the transaction.
	Before Phase = "before"
	// After represents the phase after the transaction.
	After Phase = "after"
	// Rollback represents the phase for rollback.
	Rollback Phase = "rollback"
	// Async represents the phase for asynchronous execution.
	Async Phase = "async"
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

// CreateStatement generates the Cypher statement to create the trigger.
func (t *Trigger) CreateStatement() string {
	// FIXME: Should probably use apoc.trigger.install instead of apoc.trigger.add
	return fmt.Sprintf(
		"CALL apoc.trigger.add(%s, %s, %s)",
		strconv.Quote(t.Name),
		strconv.Quote(t.KernelTransaction),
		fmt.Sprintf("{phase: %s}", strconv.Quote(string(t.Selector.Phase))),
	)
}
