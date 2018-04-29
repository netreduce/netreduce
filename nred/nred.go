/*
Package nred

TODO:
- document that it is dynamic
- document what can fail during runtime

*/
package nred

import "strconv"

type FieldType int

const (
	ConstField FieldType = iota
	GenericField
	OpaqueNumberField
	IntField
	FloatField
	StringField
	BoolField
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

type OpaqueNumber string

type Field struct {
	typ   FieldType
	name  string
	value interface{}
}

type RuleSpec struct {
	name string
	args []interface{}
}

type QuerySpec struct {
	rules []RuleSpec
}

type Definition struct {
	name    string
	value   interface{}
	queries []QuerySpec
	fields  []Field
	rules   []RuleSpec
}

var NilValue = &struct{}{}

func enumString(v int, known []string) string {
	if v < len(known) {
		return known[v]
	}

	return strconv.Itoa(v)
}

func (l FieldType) String() string { return enumString(int(l), fieldStrings) }

func Const(name string, value interface{}) Field {
	return Field{
		typ:   ConstField,
		name:  name,
		value: value,
	}
}

func Generic(name string) Field {
	return Field{
		typ:  GenericField,
		name: name,
	}
}

func Number(name string) Field {
	return Field{
		typ:  OpaqueNumberField,
		name: name,
	}
}

func Int(name string) Field {
	return Field{
		typ:  IntField,
		name: name,
	}
}

func Float(name string) Field {
	return Field{
		typ:  FloatField,
		name: name,
	}
}

func String(name string) Field {
	return Field{
		typ:  StringField,
		name: name,
	}
}

func Contains(name string, d Definition) Field {
	return Field{
		typ:   ContainsField,
		name:  name,
		value: d,
	}
}

func Rule(name string, args ...interface{}) RuleSpec {
	return RuleSpec{
		name: name,
		args: args,
	}
}

func Query(r ...RuleSpec) QuerySpec {
	return QuerySpec{
		rules: r,
	}
}

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
			d.value = at.value
		default:
			d.value = a[i]
		}
	}

	return d
}

func Export(name string, d Definition) Definition {
	d.name = name
	return d
}

func (d Definition) Name() string {
	return d.name
}

func (d Definition) Value() interface{} {
	return d.value
}
