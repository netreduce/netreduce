package data

import (
	"fmt"
	"encoding/json"
	"bytes"
	"errors"
)

type DataType int

const (
	ZeroData DataType = iota
	NumberData
	IntData
	FloatData
	StringData
	BoolData
	NilData
	StructData
	ListData
	JSONData
)

type Data struct {
	typ   DataType
	primitive interface{}
	fields map[string]Data
	items []Data
	json []byte
}

var zero Data

func Zero() Data           { return zero }
func IsZero(d Data) bool   { return d.typ == ZeroData }
func Number(n string) Data { return Data{typ: NumberData, primitive: n} }
func Int(n int64) Data     { return Data{typ: IntData, primitive: n} }
func Float(f float64) Data { return Data{typ: FloatData, primitive: f} }
func String(s string) Data { return Data{typ: StringData, primitive: s} }
func True() Data           { return Data{typ: BoolData, primitive: true} }
func False() Data          { return Data{typ: BoolData, primitive: false} }
func Nil() Data            { return Data{typ: NilData} }
func Struct(fields map[string]Data) Data         { return Data{typ: StructData, fields: fields} }
func List(items []Data) Data { return Data{typ: ListData, items: items} }
func JSON(b []byte) Data   { return Data{typ: JSONData, json: b} }

func escapeString(s string) string {
	var b []byte
	for i := range s {
		switch s[i] {
		case '\\', '"':
			b = append(b, '\\')
		}

		b = append(b, s[i])
	}

	return string(b)
}

// TODO: test loop
func (d Data) MarshalJSON() ([]byte, error) {
	switch d.typ {
	case ZeroData:
		return []byte("{}"), nil
	case StructData:
		return json.Marshal(d.fields)
	case ListData:
		return json.Marshal(d.items)
	case JSONData:
		return d.json, nil
	default:
		return json.Marshal(d.primitive)
	}
}

func (d *Data) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.UseNumber()

	if len(b) > 0 && b[0] == '{' {
		var fields map[string]Data
		if err := dec.Decode(&fields); err != nil {
			return err
		}

		if dec.More() {
			return errors.New("multiple JSON documents")
		}

		d.typ = StructData
		d.fields = fields
		return nil
	}

	if len(b) > 0 && b[0] == '[' {
		var items []Data
		if err := dec.Decode(&items); err != nil {
			return err
		}

		if dec.More() {
			return errors.New("multiple JSON documents")
		}

		d.typ = ListData
		d.items = items
		return nil
	}

	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return err
	}

	if dec.More() {
		return errors.New("multiple JSON documents")
	}

	switch vn := v.(type) {
	case json.Number:
		vi, err := vn.Int64()
		if err != nil {
			vf, err := vn.Float64()
			if err != nil {
				return err
			} else {
				d.typ = FloatData
				d.primitive = vf
			}
		} else {
			d.typ = IntData
			d.primitive = vi
		}
	case string:
		d.typ = StringData
		d.primitive = v
	default:
		return errors.New("not implemented")
	}

	return nil
}

func (d Data) String() string {
	b, err := d.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(b)
}

func mergeFields(to, from map[string]Data) {
	for n, v := range from {
		to[n] = v
	}
}

func cloneFields(f map[string]Data) map[string]Data {
	c := make(map[string]Data)
	mergeFields(c, f)
	return c
}

func appendItems(to, from []Data) []Data {
	return append(to, from...)
}

func copyItems(i []Data) []Data {
	c := make([]Data, len(i))
	copy(c, i)
	return c
}

func (d Data) SetField(name string, value Data) (Data, error) {
	// TODO: should clone the whole
	switch d.typ {
	case ZeroData, StructData:
		d.typ = StructData
	default:
		return d, fmt.Errorf("invalid data type for setting field: %s", d.typ)
	}

	d.fields = cloneFields(d.fields)
	d.fields[name] = value
	return d, nil
}

func (d Data) Merge(m ...Data) (Data, error) {
	for _, mi := range m {
		switch d.typ {
		case ZeroData:
			d = mi
		case StructData:
			switch mi.typ {
			case ZeroData:
			case StructData:
				d.fields = cloneFields(d.fields)
				mergeFields(d.fields, mi.fields)
			case JSONData:
				var mj Data
				if err := json.Unmarshal(mi.json, &mj); err != nil {
					return d, err
				}

				switch mj.typ {
				case ZeroData:
				case StructData:
					d.fields = cloneFields(d.fields)
					mergeFields(d.fields, mj.fields)
				default:
					return d, fmt.Errorf("cannot merge %v data with %v", d.typ, mj.typ)
				}
			default:
				return d, fmt.Errorf("cannot merge %v data with %v", d.typ, mi.typ)
			}
		case ListData:
			switch mi.typ {
			case ZeroData:
			case ListData:
				d.items = copyItems(d.items)
				d.items = appendItems(d.items, mi.items)
			case JSONData:
				var mj Data
				if err := json.Unmarshal(mi.json, &mj); err != nil {
					return d, err
				}

				switch mj.typ {
				case ZeroData:
				case ListData:
					d.items = copyItems(d.items)
					d.items = appendItems(d.items, mi.items)
				default:
					return d, fmt.Errorf("cannot merge %v data with %v", d.typ, mj.typ)
				}
			default:
				return d, fmt.Errorf("cannot merge %v data with %v", d.typ, mi.typ)
			}
		default:
			return d, fmt.Errorf("cannot merge %v data with %v", d.typ, mi.typ)
		}
	}

	return d, nil
}
