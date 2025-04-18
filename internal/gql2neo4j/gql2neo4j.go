package gql2neo4j

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
	"github.com/marcfrederick/gqls/internal/neo4j/triggers"
)

// ConvertString parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertString(data, neo4jVersion string, apocEnabled bool) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("error parsing GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		schemaObjects, err := Convert(graphType, neo4jVersion, apocEnabled)
		if err != nil {
			return "", fmt.Errorf("error converting graph type: %w", err)
		}
		for _, c := range schemaObjects {
			_, _ = fmt.Fprintf(&result, "%s;\n", c.CreateStatement())
		}
	}

	return result.String(), nil
}

// Convert builds Neo4j constraints from a GQL GraphType definition.
func Convert(graphType *gql.GraphType, version string, apocEnabled bool) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, nodeType := range graphType.NodeTypeSet {
		label := nodeType.NodeTypeLabelSet[0] // TODO: support multiple labels

		for _, prop := range nodeType.PropertyTypeSet {
			if c, err := buildTypeConstraint(label, prop, version, apocEnabled); err != nil {
				return nil, err
			} else if c != nil {
				schemaObjects = append(schemaObjects, c...)
			}

			if !prop.PropertyValueType.IsNullable() {
				schemaObjects = append(schemaObjects, buildExistenceConstraint(label, prop.PropertyName))
			}
		}
	}

	return schemaObjects, nil
}

// buildTypeConstraint returns a property type constraint if the Neo4j version supports it.
func buildTypeConstraint(label string, property *gql.PropertyType, version string, apocEnabled bool) ([]neo4j.SchemaObject, error) {
	supported, err := isVersionSupported(version, "5.9")
	if err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}
	if !supported {
		_, _ = fmt.Fprintf(os.Stderr, "Property type constraints unsupported in Neo4j version %s\n", version)
		return nil, nil
	}

	return createPropertyTypeConstraint(label, property, apocEnabled)
}

// buildExistenceConstraint constructs a non-null existence constraint.
func buildExistenceConstraint(label, property string) *constraints.PropertyExistenceConstraint {
	return &constraints.PropertyExistenceConstraint{
		Name:         formatConstraintName("existence", label, property),
		Pattern:      fmt.Sprintf("(n:%s)", label),
		PropertyName: fmt.Sprintf("n.%s", property),
	}
}

// createPropertyTypeConstraint maps the GQL type and builds a constraint.
func createPropertyTypeConstraint(label string, propertyType *gql.PropertyType, apocEnabled bool) ([]neo4j.SchemaObject, error) {
	switch valueType := propertyType.PropertyValueType.(type) {
	case *datatype.CharacterStringDataTypeDescriptor, *datatype.ByteStringDataTypeDescriptor:
		return []neo4j.SchemaObject{
			&constraints.PropertyTypeConstraint{
				Name:         formatConstraintName("type", label, propertyType.PropertyName),
				Pattern:      fmt.Sprintf("(n:%s)", label),
				PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
				PropertyType: string(neo4j.String),
			},
		}, nil
	case *datatype.NumericDataTypeDescriptor:
		var t neo4j.PropertyType
		switch valueType.Kind {
		case datatype.NumericKindExact:
			t = neo4j.Integer
		case datatype.NumericKindFloat:
			t = neo4j.Float
		}

		out := []neo4j.SchemaObject{
			&constraints.PropertyTypeConstraint{
				Name:         formatConstraintName("type", label, propertyType.PropertyName),
				Pattern:      fmt.Sprintf("(n:%s)", label),
				PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
				PropertyType: string(t),
			},
		}

		if apocEnabled && valueType.Precision != 64 {
			typeMin := -math.Pow(2, float64(valueType.Precision-1))
			typeMax := math.Pow(2, float64(valueType.Precision-1)) - 1
			out = append(out, &triggers.Trigger{
				Name: formatConstraintName("trigger", label, propertyType.PropertyName),
				KernelTransaction: fmt.Sprintf(
					"UNWIND $createdNodes AS node WITH node WHERE node:Person AND node.%[1]s IS NOT NULL AND (node.%[1]s < %[2]d OR node.age > %[3]d) CALL apoc.util.validate(true, '%[1]s must be between %[2]d and %[3]d', [])",
					propertyType.PropertyName,
					int64(typeMin),
					int64(typeMax),
				),
				Selector: triggers.Selector{Phase: triggers.Before},
			})
		}

		return out, nil
	case *datatype.MiscellaneousDataTypeDescriptor:
		t, err := mapMiscGQLTypeToNeo4j(valueType.Kind)
		if err != nil {
			return nil, fmt.Errorf("error mapping GQL type to Neo4j: %w", err)
		}
		return []neo4j.SchemaObject{
			&constraints.PropertyTypeConstraint{
				Name:         formatConstraintName("type", label, propertyType.PropertyName),
				Pattern:      fmt.Sprintf("(n:%s)", label),
				PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
				PropertyType: string(t),
			},
		}, nil
	}

	panic("unreachable") // enforced by go-check-sumtype
}

// formatConstraintName standardizes constraint names.
func formatConstraintName(kind, label, property string) string {
	return strings.ToLower(fmt.Sprintf("constraint_%s_%s_%s", label, property, kind))
}

// mapMiscGQLTypeToNeo4j handles miscellaneous GQL types.
func mapMiscGQLTypeToNeo4j(kind string) (neo4j.PropertyType, error) {
	switch strings.ToUpper(kind) {
	case "BOOL", "BOOLEAN":
		return neo4j.Boolean, nil
	case "DATE":
		return neo4j.Date, nil
	case "ZONED DATETIME", "TIMESTAMP WITH TIME ZONE":
		return neo4j.ZonedDateTime, nil
	case "LOCAL DATETIME", "TIMESTAMP", "TIMESTAMP WITHOUT TIME ZONE":
		return neo4j.LocalDateTime, nil
	case "ZONED TIME", "TIME WITH TIME ZONE":
		return neo4j.ZonedTime, nil
	case "LOCAL TIME", "TIME WITHOUT TIME ZONE":
		return neo4j.LocalTime, nil
	case "DURATION":
		return neo4j.Duration, nil
	default:
		return "", fmt.Errorf("error unsupported GQL type: %s", kind)
	}
}
