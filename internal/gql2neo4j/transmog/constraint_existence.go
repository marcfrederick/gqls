// Package transmog provides functionality to convert GQL schema objects into Neo4j schema objects.
package transmog

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyTransmogrifier = (*existenceConstraintTransmogrifier)(nil)

type existenceConstraintTransmogrifier struct{}

func (t *existenceConstraintTransmogrifier) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *existenceConstraintTransmogrifier) TransmogrifyProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	var pattern string
	switch entityType {
	case NodeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	case EdgeEntityType:
		pattern = fmt.Sprintf("(n:%s)", label)
	}

	var schemaObjects []neo4j.SchemaObject
	if !propertyType.PropertyValueType.IsNullable() {
		schemaObjects = append(schemaObjects, &constraints.PropertyExistenceConstraint{
			Name:         mustFormatName("existence", label, propertyType.PropertyName),
			Pattern:      pattern,
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		})
	}

	return schemaObjects, nil
}
