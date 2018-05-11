package data

import "fmt"

type DataType int

const (
	ZeroData DataType = iota
	NumberData
	IntData
	FloatData
	StringData
	BoolData
	NilData
)

type Data struct {
	typ DataType
	value interface{}
}

var zero Data

func Zero() Data { return zero }
func Number(n string) Data { return Data{typ: NumberData, value: n} }
func Int(n int) Data { return Data{typ: IntData, value: n} }
func Float(f float64) Data { return Data{typ: FloatData, value: f} }
func String(s string) Data { return Data{typ: StringData, value: s} }
func True() Data { return Data{typ: BoolData, value: true} }
func False() Data { return Data{typ: BoolData, value: false} }
func Nil() Data { return Data{typ: NilData} }

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

func (d Data) String() string {
	switch d.typ {
	case ZeroData:
		return "<zero-data>"
	case StringData:
		return fmt.Sprintf(`"%s"`, escapeString(d.value.(string)))
	default:
		return fmt.Sprint(d.value)
	}
}
