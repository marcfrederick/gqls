// Package gql2neo4j provides functionality to convert GQL schema definitions into Neo4j constraints.
package gql2neo4j

import (
	"fmt"
	"log"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

type entityType uint8

const (
	// nodeEntityType represents a node entity type.
	nodeEntityType entityType = iota
	// edgeEntityType represents an edge entity type.
	edgeEntityType
)

// ErrUnsupported is returned when a feature is not supported in the current context.
var ErrUnsupported = fmt.Errorf("unsupported")

// Opts contains options for the conversion process.
type Opts struct {
	// Neo4jVersion is the version of Neo4j to generate constraints for.
	Neo4jVersion string
	// ApocEnabled indicates if APOC is enabled in the Neo4j instance.
	ApocEnabled bool
}

// ConvertBytes parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertBytes(data []byte, opts Opts) ([]byte, error) {
	result, err := ConvertString(string(data), opts)
	if err != nil {
		return nil, fmt.Errorf("error converting GQL to Neo4j: %w", err)
	}
	return []byte(result), nil
}

// ConvertString parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertString(data string, opts Opts) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("error parsing GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		schemaObjects, err := Convert(graphType, opts)
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
func Convert(graphType *gql.GraphType, opts Opts) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, nodeType := range graphType.NodeTypeSet {
		if len(nodeType.NodeTypeLabelSet) != 1 {
			return nil, fmt.Errorf("error Neo4j only supports one label per node type: %w", ErrUnsupported)
		}
		label := nodeType.NodeTypeLabelSet[0]

		objects, err := processProperties(nodeEntityType, label, nodeType.PropertyTypeSet, opts)
		if err != nil {
			return nil, fmt.Errorf("error processing properties: %w", err)
		}
		schemaObjects = append(schemaObjects, objects...)
	}

	for _, edgeType := range graphType.EdgeTypeSet {
		if len(edgeType.EdgeTypeLabelSet) != 1 {
			return nil, fmt.Errorf("error Neo4j only supports one label per edge type: %w", ErrUnsupported)
		}
		label := edgeType.EdgeTypeLabelSet[0]

		edgeTypeObjects, err := processEdgeType(label, edgeType, opts)
		if err != nil {
			return nil, fmt.Errorf("error processing edge type: %w", err)
		}
		schemaObjects = append(schemaObjects, edgeTypeObjects...)

		propertyObjects, err := processProperties(edgeEntityType, label, edgeType.EdgeTypePropertySet, opts)
		if err != nil {
			return nil, fmt.Errorf("error processing properties: %w", err)
		}
		schemaObjects = append(schemaObjects, propertyObjects...)
	}

	return schemaObjects, nil
}

// processEdgeType processes the edge type and returns the corresponding Neo4j schema objects.
func processEdgeType(label string, edgeType *gql.EdgeType, opts Opts) ([]neo4j.SchemaObject, error) {
	if !opts.ApocEnabled {
		return []neo4j.SchemaObject{}, nil
	}

	sourceLabelSet := edgeType.SourceNodeType.NodeTypeLabelSet
	if len(sourceLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per source node type: %w", ErrUnsupported)
	}
	startNodeType := sourceLabelSet[0]

	destinationLabelSet := edgeType.DestinationNodeType.NodeTypeLabelSet
	if len(destinationLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per destination node type: %w", ErrUnsupported)
	}
	endNodeType := destinationLabelSet[0]

	edgeEndpointTrigger, err := newEdgeEndpointTrigger(label, startNodeType, endNodeType)
	if err != nil {
		return nil, fmt.Errorf("error creating edge endpoint trigger: %w", err)
	}

	return []neo4j.SchemaObject{edgeEndpointTrigger}, nil
}

// processProperties processes properties for a given label and version.
func processProperties(t entityType, label string, properties []*gql.PropertyType, opts Opts) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, prop := range properties {
		typeConstraints, err := buildTypeConstraint(t, label, prop, opts)
		if err != nil {
			return nil, fmt.Errorf("error building type constraint: %w", err)
		}
		schemaObjects = append(schemaObjects, typeConstraints...)

		if !prop.PropertyValueType.IsNullable() {
			schemaObjects = append(schemaObjects, buildExistenceConstraint(label, prop.PropertyName))
		}
	}

	return schemaObjects, nil
}

// buildTypeConstraint returns a property type constraint if the Neo4j version supports it.
func buildTypeConstraint(t entityType, label string, property *gql.PropertyType, opts Opts) ([]neo4j.SchemaObject, error) {
	supported, err := isVersionSupported(opts.Neo4jVersion, "5.9")
	if err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}
	if !supported {
		log.Printf("Property type constraints unsupported in Neo4j version %s\n", opts.Neo4jVersion)
		return nil, nil
	}

	// TODO: Technically we can create APOC triggers on older versions
	return createPropertyTypeConstraint(t, label, property, opts)
}

// buildExistenceConstraint constructs a non-null existence constraint.
func buildExistenceConstraint(label, property string) *constraints.PropertyExistenceConstraint {
	return &constraints.PropertyExistenceConstraint{
		Name:         mustFormatName("existence", label, property),
		Pattern:      fmt.Sprintf("(n:%s)", label),
		PropertyName: fmt.Sprintf("n.%s", property),
	}
}

// createPropertyTypeConstraint maps the GQL type and builds a constraint.
func createPropertyTypeConstraint(entityType entityType, label string, propertyType *gql.PropertyType, opts Opts) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	switch valueType := propertyType.PropertyValueType.(type) {
	case *datatype.CharacterStringDataTypeDescriptor:
		schemaObjects = append(schemaObjects, &constraints.PropertyTypeConstraint{
			Name:         mustFormatName("type", label, propertyType.PropertyName),
			Pattern:      fmt.Sprintf("(n:%s)", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
			PropertyType: string(neo4j.String),
		})
		if opts.ApocEnabled && (valueType.MinLength != 0 || valueType.MaxLength != 0) {
			trigger, err := newStringLengthTrigger(entityType, label, propertyType.PropertyName, valueType.MinLength, valueType.MaxLength)
			if err != nil {
				return nil, fmt.Errorf("error creating string length trigger: %w", err)
			}
			schemaObjects = append(schemaObjects, trigger)
		}
	case *datatype.NumericDataTypeDescriptor:
		if valueType.Precision >= 64 {
			return nil, fmt.Errorf("error Neo4j does not support precision >= 64 for numeric types: %w", ErrUnsupported)
		}
		if valueType.Scale != 0 {
			return nil, fmt.Errorf("error Neo4j does not support scale for numeric types: %w", ErrUnsupported)
		}

		var propertyT neo4j.PropertyType
		switch valueType.Kind {
		case datatype.NumericKindExact:
			propertyT = neo4j.Integer
		case datatype.NumericKindFloat:
			propertyT = neo4j.Float
		}

		schemaObjects = append(schemaObjects, &constraints.PropertyTypeConstraint{
			Name:         mustFormatName("type", label, propertyType.PropertyName),
			Pattern:      fmt.Sprintf("(n:%s)", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
			PropertyType: string(propertyT),
		})

		if opts.ApocEnabled && valueType.Precision != 0 && valueType.Precision != 63 {
			trigger, err := newNumericRangeTrigger(entityType, label, propertyType.PropertyName, valueType.Precision)
			if err != nil {
				return nil, fmt.Errorf("error creating numeric range trigger: %w", err)
			}
			schemaObjects = append(schemaObjects, trigger)

		}
	case *datatype.MiscellaneousDataTypeDescriptor:
		t, err := mapMiscGQLTypeToNeo4j(valueType.Kind)
		if err != nil {
			return nil, fmt.Errorf("error mapping GQL type to Neo4j: %w", err)
		}
		schemaObjects = append(schemaObjects, &constraints.PropertyTypeConstraint{
			Name:         mustFormatName("type", label, propertyType.PropertyName),
			Pattern:      fmt.Sprintf("(n:%s)", label),
			PropertyName: fmt.Sprintf("n.%s", propertyType.PropertyName),
			PropertyType: string(t),
		})
	case *datatype.ByteStringDataTypeDescriptor:
		return nil, fmt.Errorf("error Neo4j does not support byte string type: %w", ErrUnsupported)
	default:
		return nil, fmt.Errorf("error unsupported GQL type: %s", valueType)
	}

	return schemaObjects, nil
}

// mustFormatName formats the name of a constraint or trigger.
func mustFormatName(prefix string, components ...string) string {
	if len(components) == 0 {
		panic("mustFormatName requires at least one component")
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString("_")

	for i, component := range components {
		if i > 0 {
			b.WriteString("_")
		}
		b.WriteString(component)
	}

	return strings.ToLower(b.String())
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
