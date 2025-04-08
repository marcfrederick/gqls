package gql

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/marcfrederick/gqls/internal/gql/parser"
)

//go:generate antlr -Dlanguage=Go -o ./parser -Xexact-output-dir -package parser -no-visitor ./grammar/GQL.g4

// ErrNotImplemented indicates that a feature is not yet implemented.
var ErrNotImplemented = errors.New("feature not implemented")

type CreateGraphTypeListener struct {
	CreateGraphTypeStatements []*GraphType
	*parser.BaseGQLListener
}

func (l *CreateGraphTypeListener) ExitCreateGraphTypeStatement(ctx *parser.CreateGraphTypeStatementContext) {
	graphName := ctx.CatalogGraphTypeParentAndName().GetText()

	graphTypeSpec := ctx.GraphTypeSource().NestedGraphTypeSpecification()
	if graphTypeSpec == nil {
		log.Printf("Warning: No nested graph type specification found for graph type %s", graphName)
		return
	}

	elementTypeSpecs := graphTypeSpec.GraphTypeSpecificationBody().ElementTypeList().AllElementTypeSpecification()
	if elementTypeSpecs == nil {
		log.Printf("Warning: No element type specifications found for graph type %s", graphName)
		return
	}

	graphType := &GraphType{CatalogGraphParentAndName: graphName}

	for _, elementTypeSpec := range elementTypeSpecs {
		if err := l.parseElementTypeSpecification(elementTypeSpec, graphType); err != nil {
			log.Printf("Error parsing element type spec for %s: %v", graphName, err)
			return
		}
	}

	l.CreateGraphTypeStatements = append(l.CreateGraphTypeStatements, graphType)
}

func (l *CreateGraphTypeListener) parseElementTypeSpecification(elementTypeSpec parser.IElementTypeSpecificationContext, graphType *GraphType) error {
	if nodeTypeSpec := elementTypeSpec.NodeTypeSpecification(); nodeTypeSpec != nil {
		nodeType, err := l.parseNodeTypeSpecification(nodeTypeSpec)
		if err != nil {
			return fmt.Errorf("error parsing node type nodeType: %w", err)
		}
		graphType.NodeTypeSet = append(graphType.NodeTypeSet, nodeType)
	}

	if edgeTypeSpec := elementTypeSpec.EdgeTypeSpecification(); edgeTypeSpec != nil {
		edgeType, err := l.parseEdgeTypeSpecification(edgeTypeSpec)
		if err != nil {
			return fmt.Errorf("error parsing edge type edgeType: %w", err)
		}
		graphType.EdgeTypeSet = append(graphType.EdgeTypeSet, edgeType)
	}
	return nil
}

func (l *CreateGraphTypeListener) parseNodeTypeSpecification(ctx parser.INodeTypeSpecificationContext) (*NodeType, error) {
	if nodeTypePattern := ctx.NodeTypePattern(); nodeTypePattern != nil {
		return l.parseNodeTypeFiller(nodeTypePattern.NodeTypeFiller())
	} else if nodeTypePhrase := ctx.NodeTypePhrase(); nodeTypePhrase != nil {
		return l.parseNodeTypeFiller(nodeTypePhrase.NodeTypePhraseFiller().NodeTypeFiller())
	} else {
		return nil, fmt.Errorf("error missing node type specification")
	}
}

func (l *CreateGraphTypeListener) parseNodeTypeFiller(ctx parser.INodeTypeFillerContext) (*NodeType, error) {
	nodeTypePropertyTypes, err := l.parseNodeTypePropertyTypes(ctx.NodeTypeImpliedContent().NodeTypePropertyTypes())
	if err != nil {
		return nil, fmt.Errorf("error parsing node type property types: %w", err)
	}

	return &NodeType{
		NodeTypeLabelSet: l.parseLabelSetPhrase(ctx.NodeTypeImpliedContent().NodeTypeLabelSet().LabelSetPhrase()),
		PropertyTypeSet:  nodeTypePropertyTypes,
	}, nil
}

func (l *CreateGraphTypeListener) parseLabelSetPhrase(ctx parser.ILabelSetPhraseContext) []string {
	var keyLabelSet []string
	for _, label := range ctx.LabelSetSpecification().AllLabelName() {
		keyLabelSet = append(keyLabelSet, label.GetText())
	}
	return keyLabelSet
}

// parseNodeTypePropertyTypes parses the property types for a node type.
func (l *CreateGraphTypeListener) parseNodeTypePropertyTypes(ctx parser.INodeTypePropertyTypesContext) ([]*PropertyType, error) {
	if ctx == nil {
		return []*PropertyType{}, nil
	}
	return l.parsePropertyTypesSpecification(ctx.PropertyTypesSpecification())
}

func (l *CreateGraphTypeListener) parsePropertyTypesSpecification(ctx parser.IPropertyTypesSpecificationContext) ([]*PropertyType, error) {
	var propertyTypeSet []*PropertyType
	for _, propertyType := range ctx.PropertyTypeList().AllPropertyType() {
		propertyValueType, err := l.parseValueType(propertyType)
		if err != nil {
			return nil, fmt.Errorf("error parsing property value type: %w", err)
		}

		propertyTypeSet = append(propertyTypeSet, &PropertyType{
			PropertyName:      propertyType.PropertyName().GetText(),
			PropertyValueType: propertyValueType,
		})
	}
	return propertyTypeSet, nil
}

func (l *CreateGraphTypeListener) parseValueType(ctx parser.IPropertyTypeContext) (*ValueType, error) {
	valueType := ctx.PropertyValueType().GetText()
	notNull := strings.HasSuffix(valueType, "NOTNULL")
	if notNull {
		valueType = strings.TrimSuffix(valueType, "NOTNULL")
	}
	return &ValueType{
		ValueType: valueType,
		NotNull:   notNull,
	}, nil
}

func (l *CreateGraphTypeListener) parseEdgeTypeSpecification(ctx parser.IEdgeTypeSpecificationContext) (*EdgeType, error) {
	var edgeType EdgeType

	switch {
	case ctx.EdgeTypePattern() != nil:
		if err := l.parseEdgeTypePattern(ctx.EdgeTypePattern(), &edgeType); err != nil {
			return nil, fmt.Errorf("error parsing edge type pattern: %w", err)
		}
	case ctx.EdgeTypePhrase() != nil:
		return nil, fmt.Errorf("error edge type phrase not yet implemented: %w", ErrNotImplemented)
	default:
		panic("unreachable")
	}

	return &edgeType, nil
}

func (l *CreateGraphTypeListener) parseEdgeTypePattern(ctx parser.IEdgeTypePatternContext, e *EdgeType) error {
	patternDirected := ctx.EdgeTypePatternDirected()
	patternUndirected := ctx.EdgeTypePatternUndirected()

	if patternDirected != nil {
		if patternDirected.EdgeTypePatternPointingRight() != nil {
			e.Directed = true
			return l.parseEdgeTypePatternPointingRight(patternDirected.EdgeTypePatternPointingRight(), e)
		}
		if patternDirected.EdgeTypePatternPointingLeft() != nil {
			e.Directed = true
			return l.parseEdgeTypePatternPointingLeft(patternDirected.EdgeTypePatternPointingLeft(), e)
		}
	} else if patternUndirected != nil {
		e.Directed = false
		return l.parseEdgeTypePatternUndirected(patternUndirected, e)
	}

	panic("unreachable")
}

func (l *CreateGraphTypeListener) parseEdgeTypePatternPointingRight(ctx parser.IEdgeTypePatternPointingRightContext, e *EdgeType) error {
	source, err := l.parseNodeTypeFiller(ctx.SourceNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing source node type: %w", err)
	}
	e.SourceNodeType = source

	destination, err := l.parseNodeTypeFiller(ctx.DestinationNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing destination node type: %w", err)
	}
	e.DestinationNodeType = destination

	propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypePointingRight().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypePropertySet = propertySet

	return nil
}

func (l *CreateGraphTypeListener) parseEdgeTypePatternPointingLeft(ctx parser.IEdgeTypePatternPointingLeftContext, e *EdgeType) error {
	source, err := l.parseNodeTypeFiller(ctx.SourceNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing source node type: %w", err)
	}
	e.SourceNodeType = source

	destination, err := l.parseNodeTypeFiller(ctx.DestinationNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing destination node type: %w", err)
	}
	e.DestinationNodeType = destination

	propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypePointingLeft().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypePropertySet = propertySet

	return nil
}

func (l *CreateGraphTypeListener) parseEdgeTypePatternUndirected(ctx parser.IEdgeTypePatternUndirectedContext, e *EdgeType) error {
	source, err := l.parseNodeTypeFiller(ctx.SourceNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing source node type: %w", err)
	}
	e.SourceNodeType = source

	destination, err := l.parseNodeTypeFiller(ctx.DestinationNodeTypeReference().NodeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing destination node type: %w", err)
	}
	e.DestinationNodeType = destination

	propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypeUndirected().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypePropertySet = propertySet

	return nil
}

func (l *CreateGraphTypeListener) parseEdgeTypeFiller(ctx parser.IEdgeTypeFillerContext) ([]*PropertyType, error) {
	return l.parsePropertyTypesSpecification(ctx.EdgeTypeImpliedContent().EdgeTypePropertyTypes().PropertyTypesSpecification())
}
