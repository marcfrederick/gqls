package neo4j

// PropertyType represents the type of property in Neo4j.
type PropertyType string

const (
	Boolean       PropertyType = "BOOLEAN"
	Date          PropertyType = "DATE"
	Duration      PropertyType = "DURATION"
	Float         PropertyType = "FLOAT"
	Integer       PropertyType = "INTEGER"
	List          PropertyType = "LIST"
	LocalDateTime PropertyType = "LOCAL DATETIME"
	LocalTime     PropertyType = "LOCAL TIME"
	Point         PropertyType = "POINT"
	String        PropertyType = "STRING"
	ZonedDateTime PropertyType = "ZONED DATETIME"
	ZonedTime     PropertyType = "ZONED TIME"
)

// SchemaObject represents a schema object in Neo4j.
type SchemaObject interface {
	// CreateStatement generates the CREATE statement for the schema object.
	CreateStatement() string
}
