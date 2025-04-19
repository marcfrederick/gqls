package gql

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql/datatype"
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

func (l *CreateGraphTypeListener) parseValueType(ctx parser.IPropertyTypeContext) (datatype.Descriptor, error) {
	predefinedTypeLabelCtx, ok := ctx.PropertyValueType().ValueType().(*parser.PredefinedTypeLabelContext)
	if !ok {
		return nil, fmt.Errorf("error parsing value type: non-predefined types not supported")
	}

	predefinedType := predefinedTypeLabelCtx.PredefinedType()

	switch {
	case predefinedType.CharacterStringType() != nil:
		characterStringType := predefinedType.CharacterStringType()
		minLength, maxLength := -1, -1
		if characterStringType.MinLength() != nil {
			minLength, _ = strconv.Atoi(characterStringType.MinLength().GetText())
		}
		if characterStringType.MaxLength() != nil {
			maxLength, _ = strconv.Atoi(characterStringType.MaxLength().GetText())
		}
		return &datatype.CharacterStringDataTypeDescriptor{
			MinLength: minLength,
			MaxLength: maxLength,
			Nullable:  l.parseNotNull(characterStringType.NotNull()),
		}, nil
	case predefinedType.ByteStringType() != nil:
		byteStringType := predefinedType.ByteStringType()
		minLength, maxLength := -1, -1
		if byteStringType.MinLength() != nil {
			minLength, _ = strconv.Atoi(byteStringType.MinLength().GetText())
		}
		if byteStringType.MaxLength() != nil {
			maxLength, _ = strconv.Atoi(byteStringType.MaxLength().GetText())
		}
		return &datatype.ByteStringDataTypeDescriptor{
			MinLength: minLength,
			MaxLength: maxLength,
			Nullable:  l.parseNotNull(byteStringType.NotNull()),
		}, nil
	case predefinedType.NumericType() != nil:
		numericType := predefinedType.NumericType()
		switch {
		case numericType.ExactNumericType() != nil:
			switch {
			case numericType.ExactNumericType().DecimalExactNumericType() != nil:
				return l.parseDecimalExactNumericType(numericType.ExactNumericType().DecimalExactNumericType())
			case numericType.ExactNumericType().BinaryExactNumericType() != nil:
				return l.parseBinaryExactNumericType(numericType.ExactNumericType().BinaryExactNumericType())
			default:
				panic("unreachable")
			}
		case numericType.ApproximateNumericType() != nil:
			return l.parseApproximateNumericType(numericType.ApproximateNumericType())
		default:
			panic("unreachable")
		}
	default:
		valueType := ctx.PropertyValueType().GetText()
		notNull := strings.HasSuffix(valueType, "NOTNULL")
		if notNull {
			valueType = strings.TrimSuffix(valueType, "NOTNULL")
		}
		return &datatype.MiscellaneousDataTypeDescriptor{
			Kind:     valueType,
			Nullable: !notNull,
		}, nil
	}
}

func (l *CreateGraphTypeListener) parseDecimalExactNumericType(_ parser.IDecimalExactNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	return nil, fmt.Errorf("decimal exact numeric type is not supported: %w", ErrNotImplemented)
}

func (l *CreateGraphTypeListener) parseBinaryExactNumericType(ctx parser.IBinaryExactNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	switch {
	case ctx.SignedBinaryExactNumericType() != nil:
		signedBinaryExactNumericType := ctx.SignedBinaryExactNumericType()
		var precision int
		switch {
		case signedBinaryExactNumericType.INT8() != nil:
			precision = 8
		case signedBinaryExactNumericType.INT16() != nil:
			precision = 16
		case signedBinaryExactNumericType.INT32() != nil:
			precision = 32
		case signedBinaryExactNumericType.INT64() != nil:
			precision = 64
		case signedBinaryExactNumericType.INT128() != nil:
			precision = 128
		case signedBinaryExactNumericType.INT256() != nil:
			precision = 256
		default:
			// TODO: Handle REAL, FLOAT, and DOUBLE types
			parsedPrecision, err := parseOptionalInt(signedBinaryExactNumericType.Precision())
			if err != nil {
				return nil, fmt.Errorf("error parsing precision: %w", err)
			}
			precision = parsedPrecision
		}

		return &datatype.NumericDataTypeDescriptor{
			Kind:      datatype.NumericKindExact,
			Radix:     datatype.RadixBinary,
			Precision: precision,
			Scale:     -1,
			Nullable:  l.parseNotNull(signedBinaryExactNumericType.NotNull()),
		}, nil
	case ctx.UnsignedBinaryExactNumericType() != nil:
		return nil, fmt.Errorf("unsigned binary exact numeric type is not supported: %w", ErrNotImplemented)
	default:
		panic("unreachable")
	}
}

// parseApproximateNumericType parses the approximate numeric type (e.g., FLOAT, DOUBLE).
func (l *CreateGraphTypeListener) parseApproximateNumericType(ctx parser.IApproximateNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	var precision int
	switch {
	case ctx.FLOAT16() != nil:
		precision = 16
	case ctx.FLOAT32() != nil:
		precision = 32
	case ctx.FLOAT64() != nil:
		precision = 64
	case ctx.FLOAT128() != nil:
		precision = 128
	case ctx.FLOAT256() != nil:
		precision = 256
	default:
		// TODO: Handle SMALLINT, INTEGER, and BIGINT types
		parsedPrecision, err := parseOptionalInt(ctx.Precision())
		if err != nil {
			return nil, fmt.Errorf("error parsing precision: %w", err)
		}
		precision = parsedPrecision
	}

	scale, err := parseOptionalInt(ctx.Scale())
	if err != nil {
		return nil, fmt.Errorf("error parsing scale: %w", err)
	}

	return &datatype.NumericDataTypeDescriptor{
		Kind:      datatype.NumericKindFloat,
		Radix:     datatype.RadixBinary,
		Precision: precision,
		Scale:     scale,
		Nullable:  l.parseNotNull(ctx.NotNull()),
	}, nil
}

func parseOptionalInt(node antlr.ParserRuleContext) (int, error) {
	if node == nil {
		return -1, nil
	}

	// somewhat hacky, but avoids some more complex logic
	value, err := strconv.Atoi(node.GetText())
	if err != nil {
		return -1, fmt.Errorf("error parsing integer value: %w", err)
	}
	return value, nil
}

// parseNotNull parses the NOT NULL constraint from the context.
func (l *CreateGraphTypeListener) parseNotNull(ctx parser.INotNullContext) bool {
	if ctx != nil && ctx.NOT() != nil {
		return false
	}
	return true
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
	if directed := ctx.EdgeTypePatternDirected(); directed != nil {
		e.Directed = true
		if pointingRight := directed.EdgeTypePatternPointingRight(); pointingRight != nil {
			return l.parseEdgeTypePatternPointingRight(pointingRight, e)
		} else if pointingLeft := directed.EdgeTypePatternPointingLeft(); pointingLeft != nil {
			return l.parseEdgeTypePatternPointingLeft(pointingLeft, e)
		} else {
			return fmt.Errorf("error invalid directed edge type pattern: %s", ctx.GetText())
		}
	} else if undirected := ctx.EdgeTypePatternUndirected(); undirected != nil {
		e.Directed = false
		return l.parseEdgeTypePatternUndirected(undirected, e)
	} else {
		return fmt.Errorf("error invalid edge type pattern: %s", ctx.GetText())
	}
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

	labelSet, propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypePointingRight().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypeLabelSet = labelSet
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

	labelSet, propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypePointingLeft().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypeLabelSet = labelSet
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

	labelSet, propertySet, err := l.parseEdgeTypeFiller(ctx.ArcTypeUndirected().EdgeTypeFiller())
	if err != nil {
		return fmt.Errorf("error parsing property types specification: %w", err)
	}
	e.EdgeTypeLabelSet = labelSet
	e.EdgeTypePropertySet = propertySet

	return nil
}

func (l *CreateGraphTypeListener) parseEdgeTypeFiller(ctx parser.IEdgeTypeFillerContext) ([]string, []*PropertyType, error) {
	impliedContent := ctx.EdgeTypeImpliedContent()
	if impliedContent == nil {
		return nil, nil, fmt.Errorf("error: missing implied content in edge type filler")
	}

	labelSet := l.parseLabelSetPhrase(impliedContent.EdgeTypeLabelSet().LabelSetPhrase())

	var propertyTypesSpec []*PropertyType
	if edgeTypePropertyTypes := impliedContent.EdgeTypePropertyTypes(); edgeTypePropertyTypes != nil {
		var err error
		propertyTypesSpec, err = l.parsePropertyTypesSpecification(edgeTypePropertyTypes.PropertyTypesSpecification())
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing edge type property types: %w", err)
		}
	}

	return labelSet, propertyTypesSpec, nil
}
