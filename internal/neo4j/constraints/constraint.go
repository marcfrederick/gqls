// Package constraints provides a set of constraints for Neo4j.
package constraints

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/neo4j"
)

// SubjectType defines the subject of a constraint.
type SubjectType uint8

const (
	// NodeType represents a node constraint.
	NodeType SubjectType = iota
	// RelationshipType represents a relationship constraint.
	RelationshipType
)

// Constraint is a common interface for all constraints.
//
//sumtype:decl
type Constraint interface {
	// sealed is a marker to ensure that all types are sealed.
	sealed()

	neo4j.SchemaObject
}

// PropertyUniquenessConstraint ensures uniqueness.
type PropertyUniquenessConstraint struct {
	Name         string
	Pattern      string
	PropertyName string
}

var _ Constraint = (*PropertyUniquenessConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property uniqueness constraint.
func (c *PropertyUniquenessConstraint) CreateStatement() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS UNIQUE", c.Name, c.Pattern, c.PropertyName)
}

func (*PropertyUniquenessConstraint) sealed() {}

// PropertyExistenceConstraint ensures a property is not null.
type PropertyExistenceConstraint struct {
	Name         string
	Pattern      string
	PropertyName string
}

var _ Constraint = (*PropertyExistenceConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property existence constraint.
func (c *PropertyExistenceConstraint) CreateStatement() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS NOT NULL", c.Name, c.Pattern, c.PropertyName)
}

func (*PropertyExistenceConstraint) sealed() {}

// PropertyTypeConstraint enforces a specific type.
type PropertyTypeConstraint struct {
	Name         string
	Pattern      string
	PropertyName string
	PropertyType string
}

var _ Constraint = (*PropertyTypeConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property type constraint.
func (c *PropertyTypeConstraint) CreateStatement() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS :: %s", c.Name, c.Pattern, c.PropertyName, c.PropertyType)
}

func (*PropertyTypeConstraint) sealed() {}

// KeyConstraint ensures a property is a key.
type KeyConstraint struct {
	Name         string
	Pattern      string
	PropertyName string
	SubjectType  SubjectType
}

var _ Constraint = (*KeyConstraint)(nil)

// CreateStatement generates the Cypher statement to create the key constraint.
func (c *KeyConstraint) CreateStatement() string {
	switch c.SubjectType {
	case NodeType:
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS NODE KEY", c.Name, c.Pattern, c.PropertyName)
	case RelationshipType:
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS RELATIONSHIP KEY", c.Name, c.Pattern, c.PropertyName)
	}
	panic(fmt.Sprintf("invalid subject type: %d", c.SubjectType))
}

func (*KeyConstraint) sealed() {}
