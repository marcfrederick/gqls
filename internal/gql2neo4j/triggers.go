package gql2neo4j

import (
	"bytes"
	"embed"
	"fmt"
	"math"
	"path"
	"text/template"

	"github.com/marcfrederick/gqls/internal/neo4j/triggers"
)

//go:embed templates/*
var templates embed.FS

// NumericRangeTriggerData holds the data for a numeric range trigger.
type NumericRangeTriggerData struct {
	Variable     string
	Label        string
	PropertyName string
	MinValue     int
	MaxValue     int
}

// newNumericRangeTrigger creates a new numeric range trigger.
func newNumericRangeTrigger(entityType entityType, label, propertyName string, precision int) (*triggers.Trigger, error) {
	kernelTransaction, err := renderTemplate("numeric_range_trigger.cypher.tmpl", NumericRangeTriggerData{
		Variable:     entityTypeToVariable(entityType),
		Label:        label,
		PropertyName: propertyName,
		MinValue:     -int(math.Pow(2, float64(precision))),
		MaxValue:     int(math.Pow(2, float64(precision))) - 1,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := formatName("trigger", label, propertyName)
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}

// StringLengthTriggerData holds the data for a string length trigger.
type StringLengthTriggerData struct {
	Variable     string
	Label        string
	PropertyName string
	MinLength    int
	MaxLength    int
}

// newStringLengthTrigger creates a new string length trigger.
func newStringLengthTrigger(entityType entityType, label, propertyName string, minLength, maxLength int) (*triggers.Trigger, error) {
	kernelTransaction, err := renderTemplate("string_length_trigger.cypher.tmpl", StringLengthTriggerData{
		Variable:     entityTypeToVariable(entityType),
		Label:        label,
		PropertyName: propertyName,
		MinLength:    minLength,
		MaxLength:    maxLength,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := formatName("trigger", label, propertyName)
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}

func entityTypeToVariable(entityType entityType) string {
	switch entityType {
	case nodeEntityType:
		return "$createdNodes"
	case edgeEntityType:
		return "$createdRelationships"
	}
	panic("unreachable")
}

// renderTemplate renders a template with the given data.
//
// It reads the template file from the embedded filesystem, parses it, and
// executes it with the provided data.
func renderTemplate(templateName string, data any) (string, error) {
	tmplPath := path.Join("templates", templateName)

	tmplData, err := templates.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("error reading template file: %w", err)
	}

	tmpl, err := template.New(templateName).Parse(string(tmplData))
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return buf.String(), nil
}
