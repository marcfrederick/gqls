package gql

// ValueType represents a value type.
type ValueType struct {
	ValueType string
	NotNull   bool
}

// PropertyType represents a property type.
type PropertyType struct {
	PropertyName      string
	PropertyValueType *ValueType
}

// NodeType represents a node type.
type NodeType struct {
	NodeTypeLabelSet []string
	PropertyTypeSet  []*PropertyType
}

// EdgeType represents an edge type.
//
// Each edge type specification has a set of edge type labels, a set of edge
// type properties, a source node type, a destination node type, and a
// directed flag.
type EdgeType struct {
	// EdgeTypeLabelSet comprises a set of zero or more labels.
	EdgeTypeLabelSet []string
	// EdgeTypePropertySet comprises a set of zero or more properties.
	EdgeTypePropertySet []*PropertyType
	// SourceNodeType one of the endpoint node types.
	SourceNodeType *NodeType
	// DestinationNodeType the other endpoint node type.
	DestinationNodeType *NodeType
	// Directed is true if the edge is directed, false otherwise.
	Directed bool
}

// GraphType represents a graph type as a collection of node and edge type specifications.
type GraphType struct {
	CatalogGraphParentAndName string
	// NodeTypeSet is the set of node type descriptors.
	NodeTypeSet []*NodeType
	// EdgeTypeSet is the set of edge type descriptors.
	EdgeTypeSet []*EdgeType
}
