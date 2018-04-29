package nred

import (
	"fmt"
	"io"
	"bytes"
)

type writerTo interface {
	writeTo(w io.Writer) error
}

func toString(wt writerTo) string {
	b := bytes.NewBuffer(nil)
	if err := wt.writeTo(b); err != nil {
		panic(fmt.Errorf("error while printing to string: %v", err))
	}

	return b.String()
}

func fprintf(w io.Writer, f string, v interface{}) (err error) {
	if v == NilValue {
		_, err = fmt.Fprint(w, "null")
		return
	}

	switch vt := v.(type) {
	case int:
		_, err = fmt.Fprintf(w, f, vt)
	case float64:
		_, err = fmt.Fprintf(w, f, vt)
	case string:
		_, err = fmt.Fprintf(w, f, fmt.Sprintf(`"%s"`, escapeString(vt)))
	case bool:
		_, err = fmt.Fprintf(w, f, vt)
	default:
		if wt, ok := v.(writerTo); ok {
			b := bytes.NewBuffer(nil)
			if err = wt.writeTo(b); err == nil {
				_, err = fmt.Fprintf(w, f, b)
			}
		} else {
			_, err = fmt.Fprintf(w, f, v)
		}
	}

	return
}

func fprint(w io.Writer, v interface{}) (err error) {
	return fprintf(w, "%v", v)
}

func (f Field) writeTo(w io.Writer) (err error) {
	switch f.typ {
	case ConstField, ContainsField:
		_, err = fmt.Fprintf(w, `%v("%s", `, f.typ, f.name)
		err = fprint(w, f.value)
		_, err = fmt.Fprint(w, ")")
	default:
		_, err = fmt.Fprintf(w, `%v("%s")`, f.typ, f.name)
	}

	return
}

func (r RuleSpec) writeTo(w io.Writer) error {
	if r.name == urlConnectorName {
		return fprint(w, r.args[0])
	}

	if _, err := fmt.Fprintf(w, "%s(", r.name); err != nil {
		return err
	}

	if len(r.args) > 0 {
		if err := fprint(w, r.args[0]); err != nil {
			return err
		}

		rest := r.args[1:]
		for i := range rest {
			if err := fprintf(w, ", %v", rest[i]); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprint(w, ")"); err != nil {
		return err
	}

	return nil
}

func (q QuerySpec) writeTo(w io.Writer) error {
	if _, err := fmt.Fprint(w, "query("); err != nil {
		return err
	}

	if len(q.rules) > 0 {
		if err := fprint(w, q.rules[0]); err != nil {
			return err
		}

		rest := q.rules[1:]
		for i := range rest {
			if err := fprintf(w, ", %v", rest[i]); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprint(w, ")"); err != nil {
		return err
	}

	return nil
}

func (d Definition) writeTo(w io.Writer) error {
	if d.name == "" {
		if _, err := fmt.Fprint(w, "define("); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, `export "%s" define(`, d.name); err != nil {
			return err
		}
	}

	var args []interface{}
	if d.value != nil {
		args = append(args, d.value)
	}

	for i := range d.queries {
		args = append(args, d.queries[i])
	}

	for i := range d.fields {
		args = append(args, d.fields[i])
	}

	for i := range d.rules {
		args = append(args, d.rules[i])
	}

	if len(args) > 0 {
		if err := fprint(w, args[0]); err != nil {
			return err
		}

		rest := args[1:]
		for i := range rest {
			if err := fprintf(w, ", %v", rest[i]); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprint(w, ")"); err != nil {
		return err
	}

	return nil
}

func (f Field) String() string { return toString(f) }
func (r RuleSpec) String() string { return toString(r) }
func (q QuerySpec) String() string { return toString(q) }
func (d Definition) String() string { return toString(d) }

func Fprint(w io.Writer, d ...Definition) error {
	if len(d) == 0 {
		return nil
	}

	if err := fprint(w, d[0]); err != nil {
		return err
	}

	if len(d) > 1 {
		if _, err := fmt.Fprint(w, "; "); err != nil {
			return err
		}
	}

	return Fprint(w, d[1:]...)
}

func Sprint(d ...Definition) string {
	b := bytes.NewBuffer(nil)
	if err := Fprint(b, d...); err != nil {
		panic(fmt.Errorf("error while printing to string: %v", err))
	}

	return b.String()
}
