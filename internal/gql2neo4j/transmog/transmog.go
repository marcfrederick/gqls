package transmog

import (
	"embed"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/neo4j"
)

//go:embed templates/*
var triggerTemplates embed.FS

// Transmogrifiers is a list of all transmogrifiers.
var Transmogrifiers = []Transmogrifier{
	&typeConstraintTransmogrifier{},
	&uniqueConstraintTransmogrifier{},
	&existenceConstraintTransmogrifier{},
	&keyConstraintTransmogrifier{},
	&stringLengthTriggerTransmogrifier{},
	&numericRangeTriggerTransmogrifier{},
	&edgeEndpointsTriggerTransmogrifier{},
}

// Transmogrifier is an interface that defines a transmogrifier (converter).
type Transmogrifier interface {
	// IsCompatible checks if the transmogrifier is compatible with the database.
	IsCompatible(meta *DatabaseMetadata) (bool, error)
}

// NodeTransmogrifier is a transmogrifier that transmogrifies nodes.
type NodeTransmogrifier interface {
	Transmogrifier

	// TransmogrifyNode transmogrifies a node type into a list of Neo4j schema objects.
	TransmogrifyNode(nodeType *gql.NodeType) ([]neo4j.SchemaObject, error)
}

// EdgeTransmogrifier is a transmogrifier that transmogrifies edges.
type EdgeTransmogrifier interface {
	Transmogrifier

	// TransmogrifyEdge transmogrifies an edge type into a list of Neo4j schema objects.
	TransmogrifyEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error)
}

// PropertyTransmogrifier is a transmogrifier that transmogrifies properties.
type PropertyTransmogrifier interface {
	Transmogrifier

	// TransmogrifyProperty transmogrifies a property type into a list of Neo4j schema objects.
	TransmogrifyProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error)
}
