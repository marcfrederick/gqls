package gql2neo4j

import (
	"fmt"
	"os"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

// ConvertString parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertString(data, neo4jVersion string) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("error parsing GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		allConstraints, err := Convert(graphType, neo4jVersion)
		if err != nil {
			return "", fmt.Errorf("error converting graph type: %w", err)
		}
		for _, c := range allConstraints {
			_, _ = fmt.Fprintf(&result, "%s;\n", c.String())
		}
	}

	return result.String(), nil
}

// Convert builds Neo4j constraints from a GQL GraphType definition.
func Convert(graphType *gql.GraphType, version string) ([]constraints.Constraint, error) {
	var allConstraints []constraints.Constraint

	for _, nodeType := range graphType.NodeTypeSet {
		label := nodeType.NodeTypeLabelSet[0] // TODO: support multiple labels

		for _, prop := range nodeType.PropertyTypeSet {
			if c, err := buildTypeConstraint(label, prop, version); err != nil {
				return nil, err
			} else if c != nil {
				allConstraints = append(allConstraints, c)
			}

			if !prop.PropertyValueType.IsNullable() {
				allConstraints = append(allConstraints, buildExistenceConstraint(label, prop.PropertyName))
			}
		}
	}

	return allConstraints, nil
}

// buildTypeConstraint returns a property type constraint if the Neo4j version supports it.
func buildTypeConstraint(label string, property *gql.PropertyType, version string) (constraints.Constraint, error) {
	supported, err := isVersionSupported(version, "5.9")
	if err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}
	if !supported {
		_, _ = fmt.Fprintf(os.Stderr, "Property type constraints unsupported in Neo4j version %s\n", version)
		return nil, nil
	}

	return createPropertyTypeConstraint(label, property)
}

// buildExistenceConstraint constructs a non-null existence constraint.
func buildExistenceConstraint(label, property string) *constraints.PropertyExistenceConstraint {
	return &constraints.PropertyExistenceConstraint{
		BaseConstraint: constraints.BaseConstraint{
			Name:    formatConstraintName("existence", label, property),
			Pattern: fmt.Sprintf("(n:%s)", label),
		},
		PropertyName: fmt.Sprintf("n.%s", property),
	}
}

// createPropertyTypeConstraint maps the GQL type and builds a constraint.
func createPropertyTypeConstraint(label string, property *gql.PropertyType) (*constraints.PropertyTypeConstraint, error) {
	neo4jType, err := mapGQLTypeToNeo4j(property.PropertyValueType)
	if err != nil {
		return nil, err
	}
	return &constraints.PropertyTypeConstraint{
		BaseConstraint: constraints.BaseConstraint{
			Name:    formatConstraintName("type", label, property.PropertyName),
			Pattern: fmt.Sprintf("(n:%s)", label),
		},
		PropertyName: fmt.Sprintf("n.%s", property.PropertyName),
		PropertyType: string(neo4jType),
	}, nil
}

// formatConstraintName standardizes constraint names.
func formatConstraintName(kind, label, property string) string {
	return strings.ToLower(fmt.Sprintf("constraint_%s_%s_%s", label, property, kind))
}

// mapGQLTypeToNeo4j translates GQL data types to Neo4j property types.
func mapGQLTypeToNeo4j(descriptor datatype.Descriptor) (neo4j.PropertyType, error) {
	switch gqlType := descriptor.(type) {
	case *datatype.CharacterStringDataTypeDescriptor, *datatype.ByteStringDataTypeDescriptor:
		return neo4j.String, nil
	case *datatype.NumericDataTypeDescriptor:
		switch gqlType.Kind {
		case datatype.NumericKindExact:
			return neo4j.Integer, nil
		case datatype.NumericKindFloat:
			return neo4j.Float, nil
		default:
			return "", fmt.Errorf("error unexpected numeric kind: %v", gqlType.Kind)
		}
	case *datatype.MiscellaneousDataTypeDescriptor:
		return mapMiscGQLTypeToNeo4j(gqlType.Kind)
	default:
		return "", fmt.Errorf("error unsupported GQL type: %T", descriptor)
	}
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
