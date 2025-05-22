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
	if propertyType.PropertyValueType.IsNullable() {
		return nil, nil
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyExistenceConstraint{
			Name:         mustFormatName("existence", label, propertyType.PropertyName),
			Pattern:      getPattern(entityType, "n", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		},
	}, nil
}
