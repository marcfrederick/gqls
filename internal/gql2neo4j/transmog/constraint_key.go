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
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsPrimaryKey {
		return nil, nil
	}

	return []neo4j.SchemaObject{
		&constraints.KeyConstraint{
			Name:         mustFormatName("key", label, propertyType.PropertyName),
			Pattern:      getPattern(entityType, "n", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
		},
	}, nil
}
