package convert

import (
	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyConverter = (*existenceConstraintConverter)(nil)

type existenceConstraintConverter struct{}

func (t *existenceConstraintConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *existenceConstraintConverter) ConvertProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.PropertyValueType.IsNullable() {
		return nil, nil
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyExistenceConstraint{
			Name:         mustFormatName("existence", []string{label, propertyType.PropertyName}),
			Pattern:      getPattern(entityType, "n", label),
			Alias:        "n",
			PropertyName: propertyType.PropertyName,
		},
	}, nil
}
