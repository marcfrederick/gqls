package constraints_test

import (
	"testing"

	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

func TestPropertyUniquenessConstraint_CreateStatement(t *testing.T) {
	constraint := &constraints.PropertyUniquenessConstraint{
		PropertyName: "sub.some_property",
		Name:         "some_name",
		Pattern:      "(sub:Label)",
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS UNIQUE"
	if got := constraint.CreateStatement(); got != want {
		t.Errorf("CreateStatement() = %v, want %v", got, want)
	}
}

func TestPropertyExistenceConstraint_CreateStatement(t *testing.T) {
	constraint := &constraints.PropertyExistenceConstraint{
		PropertyName: "sub.some_property",
		Name:         "some_name",
		Pattern:      "(sub:Label)",
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS NOT NULL"
	if got := constraint.CreateStatement(); got != want {
		t.Errorf("CreateStatement() = %v, want %v", got, want)
	}
}

func TestPropertyTypeConstraint_CreateStatement(t *testing.T) {
	constraint := &constraints.PropertyTypeConstraint{
		PropertyName: "sub.some_property",
		PropertyType: "STRING",
		Name:         "some_name",
		Pattern:      "(sub:Label)",
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS :: STRING"
	if got := constraint.CreateStatement(); got != want {
		t.Errorf("CreateStatement() = %v, want %v", got, want)
	}
}

func TestKeyConstraint_CreateStatement(t *testing.T) {
	constraint := &constraints.KeyConstraint{
		PropertyName: "sub.some_property",
		SubjectType:  constraints.NodeType,
		Name:         "some_name",
		Pattern:      "(sub:Label)",
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS NODE KEY"
	if got := constraint.CreateStatement(); got != want {
		t.Errorf("CreateStatement() = %v, want %v", got, want)
	}
}
