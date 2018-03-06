package nred

type FieldType int

const (
	ConstantField FieldType = iota
	IntField
	StringField
	OneChildField
)

type Field struct {
	typ         FieldType
	name        string
	value       interface{}
	definitions []Definition
}

type QuerySpec struct {
	connector interface{}
}

type MapSpec struct {
	mapper func(interface{}) interface{}
}

type Definition struct {
	path string
	values []interface{}
	fields []Field
	queries  []QuerySpec
}

var ZeroQuery QuerySpec

func Define(entries ...interface{}) Definition {
	var d Definition
	for _, e := range entries {
		switch et := e.(type) {
		case Field:
			d.fields = append(d.fields, et)
		case QuerySpec:
			d.queries = append(d.queries, et)
		case int, float64, string:
			d.values = append(d.values, et)
		case Definition:
			d.fields = append(d.fields, et.fields...)
			d.queries = append(d.queries, et.queries...)
			d.values = append(d.values, et.values...)
		}
	}

	return d
}

func Extend(d Definition, entries ...interface{}) Definition {
	return Define(append(entries, d)...)
}

func Merge(d ...Definition) Definition {
	var entries []interface{}
	for i := range d {
		entries = append(entries, d[i])
	}

	return Define(entries...)
}

func Export(path string, entries ...interface{}) Definition {
	d := Define(entries...)
	d.path = path
	return d
}

func (d Definition) Path() string {
	return d.path
}

func (d Definition) Values() []interface{} {
	return d.values
}

func (d Definition) Fields() []Field {
	return d.fields
}

func (d Definition) Queries() []QuerySpec {
	return d.queries
}

func Constant(name string, value interface{}) Field {
	return Field{typ: ConstantField, name: name, value: value}
}

func Int(name string) Field {
	return Field{typ: IntField, name: name}
}

func String(name string) Field {
	return Field{typ: StringField, name: name}
}

func StringMapped(name string, mapper func(interface{}) string) Field {
	return Field{}
}

func Contains(name string, d ...Definition) Field {
	return Field{}
}

func ContainsOptional(name string, d ...Definition) Field {
	return Field{}
}

func ContainsOne(name string, d Definition) Field {
	return Field{typ: OneChildField, name: name, definitions: []Definition{d}}
}

func ContainsByKey(name string, d Definition) Field {
	return Field{}
}

func ContainsOneByKey(name string, d Definition) Field {
	return Field{}
}

func ContainsByField(field string, name string, d Definition) Field {
	return Field{}
}

func ContainsOneByField(field string, name string, d Definition) Field {
	return Field{}
}

func ContainsByFields(fields []string, name string, d Definition) Field {
	return Field{}
}

func ContainsOneByFields(fields []string, name string, d Definition) Field {
	return Field{}
}

func ContainsByFilter(name string, d Definition) Field {
	return Field{}
}

func ContainsOneByFilter(name string, d Definition) Field {
	return Field{}
}

func (f Field) Type() FieldType {
	return f.typ
}

func (f Field) Name() string {
	return f.name
}

func (f Field) Value() interface{} {
	return f.value
}

func (f Field) Definitions() []Definition {
	return f.definitions
}

func Query(connector interface{}) QuerySpec {
	return QuerySpec{connector: connector}
}

func QueryOne(connector interface{}) QuerySpec {
	return QuerySpec{connector: connector}
}

func (q QuerySpec) Connector() interface{} {
	return q.connector
}

func Map(mapper func(interface{}) interface{}) MapSpec {
	return MapSpec{mapper: mapper}
}

func (m MapSpec) Mapper() func(interface{}) interface{} {
	return m.mapper
}

func List() {}

func Normalize(d Definition) Definition { return d }
