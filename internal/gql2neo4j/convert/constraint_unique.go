package convert

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyConverter = (*uniqueConstraintConverter)(nil)

type uniqueConstraintConverter struct{}

func (t *uniqueConstraintConverter) IsCompatible(_ *DatabaseMetadata) (bool, error) {
	return true, nil
}

func (t *uniqueConstraintConverter) ConvertProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsUnique {
		return nil, nil
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyUniquenessConstraint{
			Name:         mustFormatName("unique", label, propertyType.PropertyName),
			Pattern:      getPattern(entityType, "n", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		},
	}, nil
}
