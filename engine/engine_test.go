package engine

import (
	"bytes"
	"testing"

	"github.com/netreduce/netreduce/nred"
)

func Test(t *testing.T) {
	doc := `"Hello, world!"`
	defs, err := nred.Parse(bytes.NewBufferString(doc))
	if err != nil {
		t.Fatal(err)
	}

	if len(defs) != 1 {
		t.Fatal("failed to parse query doc")
	}

	def := defs[0]
	e := New(Options{})
	c := FromQuery(def, nil)
	d, err := e.Exec(c)
	if err != nil {
		t.Fatal(err)
	}

	if d.String() != `"Hello, world!"` {
		t.Fatal("invalid data received")
	}
}
