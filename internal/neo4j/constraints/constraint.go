package constraints

import "fmt"

// ConstraintType defines different types of constraints.
type ConstraintType uint8

const (
	Unique      ConstraintType = iota // Ensures uniqueness of a property
	Existence                         // Ensures a property is not null
	TypeEnforce                       // Enforces a specific property type
	Key                               // Defines a property as a key
)

// SubjectType defines the subject of a constraint.
type SubjectType uint8

const (
	NodeType         SubjectType = iota // Represents a node constraint
	RelationshipType                    // Represents a relationship constraint
)

// Constraint is a common interface for all constraints.
type Constraint interface {
	fmt.Stringer
	ConstraintType() ConstraintType
	GetBase() *BaseConstraint
}

// BaseConstraint provides shared fields.
type BaseConstraint struct {
	Name    string
	Pattern string
}

// PropertyUniquenessConstraint ensures uniqueness.
type PropertyUniquenessConstraint struct {
	BaseConstraint
	PropertyName string
}

var _ Constraint = (*PropertyUniquenessConstraint)(nil)

func (c *PropertyUniquenessConstraint) String() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS UNIQUE", c.Name, c.Pattern, c.PropertyName)
}

func (*PropertyUniquenessConstraint) ConstraintType() ConstraintType {
	return Unique
}

func (c *PropertyUniquenessConstraint) GetBase() *BaseConstraint {
	return &c.BaseConstraint
}

// PropertyExistenceConstraint ensures a property is not null.
type PropertyExistenceConstraint struct {
	BaseConstraint
	PropertyName string
}

var _ Constraint = (*PropertyExistenceConstraint)(nil)

func (c *PropertyExistenceConstraint) String() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS NOT NULL", c.Name, c.Pattern, c.PropertyName)
}

func (*PropertyExistenceConstraint) ConstraintType() ConstraintType {
	return Existence
}

func (c *PropertyExistenceConstraint) GetBase() *BaseConstraint {
	return &c.BaseConstraint
}

// PropertyTypeConstraint enforces a specific type.
type PropertyTypeConstraint struct {
	BaseConstraint
	PropertyName string
	PropertyType string
}

var _ Constraint = (*PropertyTypeConstraint)(nil)

func (c *PropertyTypeConstraint) String() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS :: %s", c.Name, c.Pattern, c.PropertyName, c.PropertyType)
}

func (*PropertyTypeConstraint) ConstraintType() ConstraintType {
	return TypeEnforce
}

func (c *PropertyTypeConstraint) GetBase() *BaseConstraint {
	return &c.BaseConstraint
}

// KeyConstraint ensures a property is a key.
type KeyConstraint struct {
	BaseConstraint
	PropertyName string
	SubjectType  SubjectType
}

var _ Constraint = (*KeyConstraint)(nil)

func (c *KeyConstraint) String() string {
	switch c.SubjectType {
	case NodeType:
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS NODE KEY", c.Name, c.Pattern, c.PropertyName)
	case RelationshipType:
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS RELATIONSHIP KEY", c.Name, c.Pattern, c.PropertyName)
	default:
		panic(fmt.Sprintf("invalid subject type: %d", c.SubjectType))
	}
}

func (*KeyConstraint) ConstraintType() ConstraintType {
	return Key
}

func (c *KeyConstraint) GetBase() *BaseConstraint {
	return &c.BaseConstraint
}
