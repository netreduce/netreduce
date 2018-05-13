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
	value data.Data
	contains Definition
}

// Rule represents a rule in a definition or a query. E.g. the URL of a simple http GET query.
type Rule struct {
	name string
	args []interface{}
}

// Query represents a query for backend data. Its actual behavior depends on the contained rules.
type Query struct {
	rules []Rule
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
	queries []Query
	fields  []Field
	rules   []Rule
}

// String returns the string name associated with the FieldTyep. It's also used as the field declaration in the
// nred DSL.
func (l FieldType) String() string { return enumString(int(l), fieldStrings) }

// Const declares a const field.
func Const(name string, value data.Data) Field {
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
		contains: d,
	}
}

// NewRule declares a rule either for a query or a definition.
func NewRule(name string, args ...interface{}) Rule {
	return Rule{
		name: name,
		args: args,
	}
}

// NewQuery declares a query for a definition.
func NewQuery(r ...Rule) Query {
	return Query{
		rules: r,
	}
}

func (d Definition) Query(q ...Query) Definition {
	d.queries = append(d.queries, q...)
	return d
}

func (d Definition) Field(f ...Field) Definition {
	d.fields = append(d.fields, f...)
	return d
}

func (d Definition) Rule(r ...Rule) Definition {
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
		if !data.IsZero(ei.value) {
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

func (d Definition) Fields() []Field {
	return d.fields
}

func (d Definition) Queries() []Query {
	return d.queries
}

func (q Query) Rules() []Rule {
	return q.rules
}

func (r Rule) Name() string { return r.name }
func (r Rule) Args() []interface{} { return r.args }

func (f Field) Type() FieldType { return f.typ }
func (f Field) Name() string { return f.name }
func (f Field) Value() data.Data { return f.value }
