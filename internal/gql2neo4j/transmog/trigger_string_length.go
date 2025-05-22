package transmog

import (
	"fmt"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/triggers"
)

var _ PropertyTransmogrifier = (*stringLengthTriggerTransmogrifier)(nil)

type stringLengthTriggerTransmogrifier struct{}

func (p *stringLengthTriggerTransmogrifier) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.APOCEnabled, nil
}

func (p *stringLengthTriggerTransmogrifier) TransmogrifyProperty(entityType EntityType, label string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	typeDesc, ok := propertyType.PropertyValueType.(*datatype.CharacterStringDataTypeDescriptor)
	if !ok {
		return nil, nil
	}
	if typeDesc.MinLength == 0 && typeDesc.MaxLength == 0 {
		return nil, nil
	}

	trigger, err := newStringLengthTrigger(entityType, label, propertyType.PropertyName, typeDesc.MinLength, typeDesc.MaxLength)
	if err != nil {
		return nil, fmt.Errorf("error creating string length trigger: %w", err)
	}
	return []neo4j.SchemaObject{trigger}, nil
}

// StringLengthTriggerData holds the data for a string length trigger.
type StringLengthTriggerData struct {
	CreatedVariable            string
	AssignedPropertiesVariable string
	Label                      string
	PropertyName               string
	MinLength                  int
	MaxLength                  int
}

// newStringLengthTrigger creates a new string length trigger.
func newStringLengthTrigger(entityType EntityType, label, propertyName string, minLength, maxLength int) (*triggers.Trigger, error) {
	createdVariable, assignedPropertiesVariable := entityTypeToVariables(entityType)
	kernelTransaction, err := renderTemplate("string_length_trigger.cypher.tmpl", StringLengthTriggerData{
		CreatedVariable:            createdVariable,
		AssignedPropertiesVariable: assignedPropertiesVariable,
		Label:                      label,
		PropertyName:               propertyName,
		MinLength:                  minLength,
		MaxLength:                  maxLength,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("string_length", label, propertyName)
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}
