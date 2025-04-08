package constraints_test

import (
	"testing"

	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

func TestPropertyUniquenessConstraint_String(t *testing.T) {
	constraint := &constraints.PropertyUniquenessConstraint{
		PropertyName: "sub.some_property",
		BaseConstraint: constraints.BaseConstraint{
			Name:    "some_name",
			Pattern: "(sub:Label)",
		},
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS UNIQUE"
	if got := constraint.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestPropertyExistenceConstraint_String(t *testing.T) {
	constraint := &constraints.PropertyExistenceConstraint{
		PropertyName: "sub.some_property",
		BaseConstraint: constraints.BaseConstraint{
			Name:    "some_name",
			Pattern: "(sub:Label)",
		},
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS NOT NULL"
	if got := constraint.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestPropertyTypeConstraint_String(t *testing.T) {
	constraint := &constraints.PropertyTypeConstraint{
		PropertyName: "sub.some_property",
		PropertyType: "STRING",
		BaseConstraint: constraints.BaseConstraint{
			Name:    "some_name",
			Pattern: "(sub:Label)",
		},
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS :: STRING"
	if got := constraint.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestKeyConstraint_String(t *testing.T) {
	constraint := &constraints.KeyConstraint{
		PropertyName: "sub.some_property",
		SubjectType:  constraints.NodeType,
		BaseConstraint: constraints.BaseConstraint{
			Name:    "some_name",
			Pattern: "(sub:Label)",
		},
	}

	want := "CREATE CONSTRAINT some_name FOR (sub:Label) REQUIRE sub.some_property IS NODE KEY"
	if got := constraint.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}
