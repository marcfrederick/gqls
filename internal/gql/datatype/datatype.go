package datatype

// Descriptor describes a GQL data type.
// Based on chapter 4.16 of the GQL standard.
//
//sumtype:decl
type Descriptor interface {
	// IsNullable returns true if the data type is nullable, false otherwise.
	IsNullable() bool

	// sealed is a marker to ensure that all types are sealed.
	sealed()
}

// --- CHARACTER STRING ---

// CharacterStringDataTypeDescriptor describes a CHAR or VARCHAR data type.
type CharacterStringDataTypeDescriptor struct {
	MinLength int
	MaxLength int
	Nullable  bool
}

var _ Descriptor = (*CharacterStringDataTypeDescriptor)(nil)

func (c *CharacterStringDataTypeDescriptor) IsNullable() bool { return c.Nullable }
func (c *CharacterStringDataTypeDescriptor) sealed()          {}

// --- BYTE STRING ---

// ByteStringDataTypeDescriptor describes a BINARY or VARBINARY data type.
type ByteStringDataTypeDescriptor struct {
	MinLength int
	MaxLength int
	Nullable  bool
}

var _ Descriptor = (*ByteStringDataTypeDescriptor)(nil)

func (b *ByteStringDataTypeDescriptor) IsNullable() bool { return b.Nullable }
func (b *ByteStringDataTypeDescriptor) sealed()          {}

// --- NUMERIC ---

// NumericKind defines the kind of numeric data.
type NumericKind string

const (
	NumericKindExact NumericKind = "EXACT NUMERIC DATA"
	NumericKindFloat NumericKind = "FLOAT NUMERIC DATA"
)

// Radix defines the numeric base used.
type Radix string

const (
	RadixBinary  Radix = "BINARY"
	RadixDecimal Radix = "DECIMAL"
)

// NumericDataTypeDescriptor describes a numeric data type.
type NumericDataTypeDescriptor struct {
	Kind      NumericKind
	Radix     Radix
	Precision int // > 0
	Scale     int // >= 0
	Nullable  bool
}

var _ Descriptor = (*NumericDataTypeDescriptor)(nil)

func (n *NumericDataTypeDescriptor) IsNullable() bool { return n.Nullable }
func (n *NumericDataTypeDescriptor) sealed()          {}

// --- MISCELLANEOUS ---

// MiscellaneousDataTypeDescriptor describes a miscellaneous data type.
type MiscellaneousDataTypeDescriptor struct {
	Kind     string
	Nullable bool
}

var _ Descriptor = (*MiscellaneousDataTypeDescriptor)(nil)

func (m *MiscellaneousDataTypeDescriptor) IsNullable() bool { return m.Nullable }
func (m *MiscellaneousDataTypeDescriptor) sealed()          {}
