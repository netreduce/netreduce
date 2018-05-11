/*
Package nred provides the endpoint definitions in form of in-memory objects and a definition language parser for
Netreduce endpoints.
*/
package nred

import "github.com/netreduce/netreduce/data"

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
//
// The contents of the Definition can be set with the Query, Rule, Field, Value and Extend methods, in a union
// style.
//
// Queries are used to propagate the fields of the definition, or if no fields are set, then the result of the
// query will be used as is. Fields describe the shape of the data returned by a definition, and they filter the
// response of the queries. Rules may apply changes to the data and fields returned by the definition. The value
// defines the result of the definition alone. Currently, it is undefined what a definition will return when
// both a value and a combination of fields and queries are defined, it may become an invalid definition.
//
type Definition struct {
	name    string
	value   data.Data
	queries []QuerySpec
	fields  []Field
	rules   []RuleSpec
}

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

func (d Definition) Query(q ...QuerySpec) Definition {
	d.queries = append(d.queries, q...)
	return d
}

func (d Definition) Field(f ...Field) Definition {
	d.fields = append(d.fields, f...)
	return d
}

func (d Definition) Rule(r ...RuleSpec) Definition {
	d.rules = append(d.rules, r...)
	return d
}

func (d Definition) SetValue(v data.Data) Definition {
	d.value = v
	return d
}

func (d Definition) Extend(e ...Definition) Definition {
	for _, ei := range e {
		d.queries = append(d.queries, ei.queries...)
		d.fields = append(d.fields, ei.fields...)
		d.rules = append(d.rules, ei.rules...)
		if ei.value != data.Zero() {
			d.value = ei.value
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

func (d Definition) GetValue() data.Data {
	return d.value
}
