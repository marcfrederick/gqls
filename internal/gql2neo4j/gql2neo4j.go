// Package gql2neo4j provides functionality to convert GQL schema definitions into Neo4j constraints.
package gql2neo4j

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/parser"
	"github.com/marcfrederick/gqls/internal/gql2neo4j/transmog"
	"github.com/marcfrederick/gqls/internal/neo4j"
)

// ErrUnsupported is returned when a feature is not supported in the current context.
var ErrUnsupported = fmt.Errorf("unsupported")

// ConvertBytes parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertBytes(data []byte, meta transmog.DatabaseMetadata) ([]byte, error) {
	result, err := ConvertString(string(data), meta)
	if err != nil {
		return nil, fmt.Errorf("error converting GQL to Neo4j: %w", err)
	}
	return []byte(result), nil
}

// ConvertString parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertString(data string, meta transmog.DatabaseMetadata) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("error parsing GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		schemaObjects, err := Convert(graphType, meta)
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
func Convert(graphType *gql.GraphType, meta transmog.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, nodeType := range graphType.NodeTypeSet {
		if len(nodeType.NodeTypeLabelSet) != 1 {
			return nil, fmt.Errorf("error Neo4j only supports one label per node type: %w", ErrUnsupported)
		}
		label := nodeType.NodeTypeLabelSet[0]

		nodeSchemaObjects, err := transmogrifyNode(nodeType, meta)
		if err != nil {
			return nil, fmt.Errorf("error transmogrifying node type: %w", err)
		}
		schemaObjects = append(schemaObjects, nodeSchemaObjects...)

		objects, err := processProperties(transmog.Transmogrifiers, transmog.NodeEntityType, label, nodeType.PropertyTypeSet, &meta)
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

		edgeSchemaObjects, err := transmogrifyEdge(edgeType, meta)
		if err != nil {
			return nil, fmt.Errorf("error transmogrifying edge type: %w", err)
		}
		schemaObjects = append(schemaObjects, edgeSchemaObjects...)

		propertyObjects, err := processProperties(transmog.Transmogrifiers, transmog.EdgeEntityType, label, edgeType.EdgeTypePropertySet, &meta)
		if err != nil {
			return nil, fmt.Errorf("error processing properties: %w", err)
		}
		schemaObjects = append(schemaObjects, propertyObjects...)
	}

	return schemaObjects, nil
}

// processProperties processes properties for a given label and version.
func processProperties(transmogrifiers []transmog.Transmogrifier, entityType transmog.EntityType, label string, properties []*gql.PropertyType, meta *transmog.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, property := range properties {
		for _, transmogrifier := range transmogrifiers {
			propertyTransmogrifier, ok := transmogrifier.(transmog.PropertyTransmogrifier)
			if !ok {
				continue
			}

			isCompatible, err := propertyTransmogrifier.IsCompatible(meta)
			if err != nil {
				return nil, fmt.Errorf("error checking compatibility: %w", err)
			}
			if !isCompatible {
				continue
			}

			transmogrifiedObjects, err := propertyTransmogrifier.TransmogrifyProperty(entityType, label, property)
			if err != nil {
				return nil, fmt.Errorf("error transmogrifying property: %w", err)
			}
			schemaObjects = append(schemaObjects, transmogrifiedObjects...)
		}
	}

	return schemaObjects, nil
}

func transmogrifyNode(nodeType *gql.NodeType, meta transmog.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject
	for _, transmogrifier := range transmog.Transmogrifiers {
		nodeTransmogrifier, ok := transmogrifier.(transmog.NodeTransmogrifier)
		if !ok {
			continue
		}

		isCompatible, err := nodeTransmogrifier.IsCompatible(&meta)
		if err != nil {
			return nil, fmt.Errorf("error checking compatibility: %w", err)
		}

		if !isCompatible {
			continue
		}

		transmogrifiedObjects, err := nodeTransmogrifier.TransmogrifyNode(nodeType)
		if err != nil {
			return nil, fmt.Errorf("error transmogrifying node type: %w", err)
		}
		schemaObjects = append(schemaObjects, transmogrifiedObjects...)
	}
	return schemaObjects, nil
}

func transmogrifyEdge(edgeType *gql.EdgeType, meta transmog.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject
	for _, transmogrifier := range transmog.Transmogrifiers {
		nodeTransmogrifier, ok := transmogrifier.(transmog.EdgeTransmogrifier)
		if !ok {
			continue
		}

		isCompatible, err := nodeTransmogrifier.IsCompatible(&meta)
		if err != nil {
			return nil, fmt.Errorf("error checking compatibility: %w", err)
		}

		if !isCompatible {
			continue
		}

		transmogrifiedObjects, err := nodeTransmogrifier.TransmogrifyEdge(edgeType)
		if err != nil {
			return nil, fmt.Errorf("error transmogrifying edge type: %w", err)
		}
		schemaObjects = append(schemaObjects, transmogrifiedObjects...)
	}
	return schemaObjects, nil
}

// parseGQL parses a GQL input string into a listener structure.
func parseGQL(input string) (*gql.CreateGraphStatementListener, error) {
	lexer := parser.NewGQLLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewGQLParser(stream)

	errorListener := &gql.ErrorListener{}
	p.AddErrorListener(errorListener)

	createGraphTypeListener := &gql.CreateGraphStatementListener{}
	antlr.ParseTreeWalkerDefault.Walk(createGraphTypeListener, p.GqlProgram())

	if len(errorListener.SyntaxErrors) > 0 {
		return nil, fmt.Errorf("syntax errors: %v", errorListener.SyntaxErrors)
	}

	return createGraphTypeListener, nil
}
