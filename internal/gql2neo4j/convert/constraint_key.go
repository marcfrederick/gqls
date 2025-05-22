package convert

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

var _ PropertyConverter = (*keyConstraintConverter)(nil)

type keyConstraintConverter struct{}

func (t *keyConstraintConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *keyConstraintConverter) ConvertProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsPrimaryKey {
		return nil, nil
	}

	subjectType, ok := mapSubjectType(entityType)
	if !ok {
		return nil, fmt.Errorf("unsupported entity type for key constraint")
	}

	return []neo4j.SchemaObject{
		&constraints.KeyConstraint{
			Name:         mustFormatName("key", label, propertyType.PropertyName),
			Pattern:      getPattern(entityType, "n", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
			SubjectType:  subjectType,
		},
	}, nil
}

func mapSubjectType(entityType EntityType) (constraints.SubjectType, bool) {
	switch entityType {
	case NodeEntityType:
		return constraints.NodeType, true
	case EdgeEntityType:
		return constraints.RelationshipType, true
	default:
		return constraints.SubjectType(0), false
	}
}
