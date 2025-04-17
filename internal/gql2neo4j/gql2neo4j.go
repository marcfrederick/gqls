package gql2neo4j

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/gql/parser"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

// Convert parses GQL schema data and generates Neo4j constraints based on the provided Neo4j version.
func Convert(data, neo4jVersion string) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("failed to parse GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		constraints, err := GenerateConstraints(graphType, neo4jVersion)
		if err != nil {
			return "", fmt.Errorf("error generating constraints: %w", err)
		}
		for _, c := range constraints {
			result.WriteString(c.String())
			result.WriteString(";\n")
		}
	}
	return result.String(), nil
}

// parseGQL uses ANTLR to parse the GQL source into graph type statements.
func parseGQL(input string) (*gql.CreateGraphTypeListener, error) {
	is := antlr.NewInputStream(input)
	lexer := parser.NewGQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewGQLParser(stream)

	listener := &gql.CreateGraphTypeListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, parser.GqlProgram())
	return listener, nil
}

// GenerateConstraints builds all applicable constraints for the given GraphType.
func GenerateConstraints(graphType *gql.GraphType, version string) ([]constraints.Constraint, error) {
	var result []constraints.Constraint

	for _, nodeType := range graphType.NodeTypeSet {
		nodeLabel := nodeType.NodeTypeLabelSet[0] // TODO: support multiple labels?

		for _, property := range nodeType.PropertyTypeSet {
			constraint, err := buildPropertyTypeConstraintIfSupported(nodeLabel, property, version)
			if err != nil {
				return nil, err
			}
			if constraint != nil {
				result = append(result, constraint)
			}

			if !property.PropertyValueType.IsNullable() {
				result = append(result, createExistenceConstraint(nodeLabel, property.PropertyName))
			}
		}
	}

	return result, nil
}

func buildPropertyTypeConstraintIfSupported(label string, property *gql.PropertyType, version string) (constraints.Constraint, error) {
	supported, err := isVersionSupported(version, "5.9")
	if err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}
	if !supported {
		_, err := fmt.Fprintf(os.Stderr, "Property type constraints are not supported in Neo4j version %s\n", version)
		if err != nil {
			return nil, fmt.Errorf("failed to write to stderr: %w", err)
		}
		return nil, nil
	}

	return buildPropertyTypeConstraint(label, property)
}

func createExistenceConstraint(label, property string) *constraints.PropertyExistenceConstraint {
	return &constraints.PropertyExistenceConstraint{
		BaseConstraint: constraints.BaseConstraint{
			Name:    formatConstraintName("existence", label, property),
			Pattern: fmt.Sprintf("(n:%s)", label),
		},
		PropertyName: fmt.Sprintf("n.%s", property),
	}
}

func buildPropertyTypeConstraint(label string, property *gql.PropertyType) (*constraints.PropertyTypeConstraint, error) {
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

func formatConstraintName(kind, label, property string) string {
	return strings.ToLower(fmt.Sprintf("constraint_%s_%s_%s", label, property, kind))
}

// mapGQLTypeToNeo4j maps a GQL type to a Neo4j property type.
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
			panic("unreachable")
		}
	case *datatype.MiscellaneousDataTypeDescriptor:
		switch strings.ToUpper(gqlType.Kind) {
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
			return "", fmt.Errorf("unsupported GQL type: %s", gqlType.Kind)
		}
	default:
		panic("unreachable")
	}
}

// isVersionSupported checks if the current version is greater than or equal to the minimum required version.
func isVersionSupported(current, minimum string) (bool, error) {
	if current == "" {
		return true, nil
	}

	currVer, err := semver.NewVersion(current)
	if err != nil {
		return false, fmt.Errorf("invalid version: %s", current)
	}
	minVer, err := semver.NewVersion(minimum)
	if err != nil {
		return false, fmt.Errorf("invalid required version: %s", minimum)
	}
	return currVer.GreaterThanEqual(minVer), nil
}
