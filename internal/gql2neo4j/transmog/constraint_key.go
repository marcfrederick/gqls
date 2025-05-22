package transmog

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyTransmogrifier = (*keyConstraintTransmogrifier)(nil)

type keyConstraintTransmogrifier struct{}

func (t *keyConstraintTransmogrifier) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *keyConstraintTransmogrifier) TransmogrifyProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	var pattern string
	switch entityType {
	case NodeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	case EdgeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	}

	var schemaObjects []neo4j.SchemaObject
	if propertyType.ExtendedAttributes != nil && propertyType.ExtendedAttributes.IsPrimaryKey {
		schemaObjects = append(schemaObjects, &constraints.KeyConstraint{
			Name:         mustFormatName("key", label, propertyType.PropertyName),
			Pattern:      pattern,
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		})
	}

	return schemaObjects, nil
}
