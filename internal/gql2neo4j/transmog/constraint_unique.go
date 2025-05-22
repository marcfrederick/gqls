package transmog

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyTransmogrifier = (*uniqueConstraintTransmogrifier)(nil)

type uniqueConstraintTransmogrifier struct{}

func (t *uniqueConstraintTransmogrifier) IsCompatible(_ *DatabaseMetadata) (bool, error) {
	return true, nil
}

func (t *uniqueConstraintTransmogrifier) TransmogrifyProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	var pattern string
	switch entityType {
	case NodeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	case EdgeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	}

	var schemaObjects []neo4j.SchemaObject
	if propertyType.ExtendedAttributes != nil && propertyType.ExtendedAttributes.IsUnique {
		schemaObjects = append(schemaObjects, &constraints.PropertyUniquenessConstraint{
			Name:         mustFormatName("unique", label, propertyType.PropertyName),
			Pattern:      pattern,
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		})
	}

	return schemaObjects, nil
}
