/*
Package nred provides the endpoint definitions in form of in-memory objects and a definition language parser for
Netreduce endpoints.
*/
package nred

// FieldType defines the possible field types in nred definitions.
type FieldType int

const (

	// ConstField identifies a field with a name and a constant value.
	ConstField FieldType = iota

	// GenericField identifies a field with a name and any arbitrary value as receieved from the backend
	// (or an underlying definition or rule).
	GenericField

	// OpaqueNumberField identifies a field with a numeric value that is not parsed by Netreduce, and it
	// can be float, decimal or integer, and of arbitrary size.
	OpaqueNumberField

	// IntField identifies a field with an integer value, whose size depends on the compiler architecture.
	// Probably 64 bit.
	IntField

	// FloatField identifies a field with a 64 bit floating point numeric value.
	FloatField

	// StringField identifies a field with a string value.
	StringField

	// BoolField identifies a field with a value of either true or false.
	BoolField

	// ContainsField identifies a field in a definition that contains another definition.
	ContainsField
)

var fieldStrings = []string{
	"const",
	"generic",
	"number",
	"int",
	"float",
	"string",
	"bool",
	"contains",
}

// OpaqueNumber represents an unparsed numeric value, either int, float or decimal, and of any arbitrary size.
type OpaqueNumber string

// Field represents a field in a definition.
type Field struct {
	typ   FieldType
	name  string
	value interface{}
}

// RuleSpec represents a rule in a definition or a query. E.g. the URL of a simple http GET query.
type RuleSpec struct {
	name string
	args []interface{}
}

// QuerySpec represents a query for backend data. Its actual behavior depends on the contained rules.
type QuerySpec struct {
	rules []RuleSpec
}

// Definition describes a Netreduce request handler. Definitions may contain further Definitions. A definition
// can have a single scalar value or a set of fields. The value or the fields are populated from the responses
// of the contained queries, typically responses from requests made to backend services. The value or the fields
// can be altered, or even generated, by the contained rules.
type Definition struct {
	name    string
	value   interface{}
	queries []QuerySpec
	fields  []Field
	rules   []RuleSpec
}

// NilValue represents the 'null' data. E.g. the JSON null value.
var NilValue = &struct{}{}

// String returns the string name associated with the FieldTyep. It's also used as the field declaration in the
// nred DSL.
func (l FieldType) String() string { return enumString(int(l), fieldStrings) }

// Const declares a const field.
func Const(name string, value interface{}) Field {
	return Field{
		typ:   ConstField,
		name:  name,
		value: value,
	}
}

// Generic declares a generic field whose value type is depends on the response of the used query.
func Generic(name string) Field {
	return Field{
		typ:  GenericField,
		name: name,
	}
}

// Number declares an opaque number field.
func Number(name string) Field {
	return Field{
		typ:  OpaqueNumberField,
		name: name,
	}
}

// Int declares an int field.
func Int(name string) Field {
	return Field{
		typ:  IntField,
		name: name,
	}
}

// Float declares a floating point field.
func Float(name string) Field {
	return Field{
		typ:  FloatField,
		name: name,
	}
}

// String declares a string field.
func String(name string) Field {
	return Field{
		typ:  StringField,
		name: name,
	}
}

// Contains declares a field that contains a child definition.
func Contains(name string, d Definition) Field {
	return Field{
		typ:   ContainsField,
		name:  name,
		value: d,
	}
}

// Rule declares a rule either for a query or a definition.
func Rule(name string, args ...interface{}) RuleSpec {
	return RuleSpec{
		name: name,
		args: args,
	}
}

// Query declares a query for a definition.
func Query(r ...RuleSpec) QuerySpec {
	return QuerySpec{
		rules: r,
	}
}

// Define declares a Netreduce definition. The arguments may be of the type QuerySpec, Field, RuleSpec,
// Definition, or any arbitrary value. When an argument is a query, it will be used to propagate the fields of
// the definition, or if no fields were defined, the result of the query will be used as received. Fields
// describe the shape of the data returned by a definition, and filter the response received from the queries.
// Rules apply changes to the data and fields returned by the definition. Definition arguments are merged into
// the new definition forming a union. Values passed in as arguments to a new definition define the response of
// a definition alone. Only the last value is considered as the value of the definition and the rest is ignored.
// It is undefined what a definition will return when both a value or a combination of fields and queries are
// defined. Netreduce will do best effort to serialize those values in the response that are not plain data
// objects.
func Define(a ...interface{}) Definition {
	var d Definition
	for i := range a {
		switch at := a[i].(type) {
		case QuerySpec:
			d.queries = append(d.queries, at)
		case Field:
			d.fields = append(d.fields, at)
		case RuleSpec:
			d.rules = append(d.rules, at)
		case Definition:
			d.queries = append(d.queries, at.queries...)
			d.fields = append(d.fields, at.fields...)
			d.rules = append(d.rules, at.rules...)
			if at.value != nil {
				d.value = at.value
			}
		default:
			d.value = a[i]
		}
	}

	return d
}

// Export declares a standalone exported (not contained or shared) definition. Only exported definitions can
// represent a Netreduce endpoint.
func Export(name string, d Definition) Definition {
	d.name = name
	return d
}

// Name returns the exported name of a definition.
func (d Definition) Name() string {
	return d.name
}
