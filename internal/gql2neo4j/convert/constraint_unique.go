package convert

import (
	"fmt"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var (
	_ NodeConverter     = (*uniqueConstraintConverter)(nil)
	_ EdgeConverter     = (*uniqueConstraintConverter)(nil)
	_ PropertyConverter = (*uniqueConstraintConverter)(nil)
)

type uniqueConstraintConverter struct{}

func (t *uniqueConstraintConverter) IsCompatible(_ *DatabaseMetadata) (bool, error) {
	return true, nil
}

func (t *uniqueConstraintConverter) ConvertNode(nodeType *gql.NodeType) ([]neo4j.SchemaObject, error) {
	if nodeType.ExtendedAttributes == nil || len(nodeType.ExtendedAttributes.CompositePrimaryKeyFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(nodeType.NodeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for node type %s: %w", nodeType.NodeTypeLabelSet, err)
	}

	return newUniqueConstraint(NodeEntityType, label, nodeType.ExtendedAttributes.CompositePrimaryKeyFields)
}

func (t *uniqueConstraintConverter) ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error) {
	if edgeType.ExtendedAttributes == nil || len(edgeType.ExtendedAttributes.CompositePrimaryKeyFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(edgeType.EdgeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for edge type %s: %w", edgeType.EdgeTypeLabelSet, err)
	}

	return newUniqueConstraint(EdgeEntityType, label, edgeType.ExtendedAttributes.CompositePrimaryKeyFields)
}

func (t *uniqueConstraintConverter) ConvertProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsUnique {
		return nil, nil
	}

	return newUniqueConstraint(entityType, label, []string{propertyType.PropertyName})
}

func newUniqueConstraint(entityType EntityType, label string, fields []string) ([]neo4j.SchemaObject, error) {
	qualifiedFields := make([]string, len(fields))
	for i, field := range fields {
		qualifiedFields[i] = fmt.Sprintf("n.%s", field)
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyUniquenessConstraint{
			Name:         mustFormatName("unique", fields...),
			Pattern:      getPattern(entityType, "n", label),
			PropertyName: fmt.Sprintf("(%s)", strings.Join(qualifiedFields, ", ")),
		},
	}, nil
}
