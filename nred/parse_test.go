package nred

import (
	"bytes"
	"testing"

	"github.com/netreduce/netreduce/registry"
)

func TestParse(t *testing.T) {
	t.Run("constant", func(t *testing.T) {
		const code = `export "foo" "foo"`

		buf := bytes.NewBufferString(code)
		d, err := Parse(&registry.Registry{}, buf)
		if err != nil {
			t.Error(err)
			return
		}

		if len(d) != 1 {
			t.Error("invalid number of definitions")
			return
		}

		if d[0].Path() != "foo" {
			t.Error("failed to parse path")
			return
		}

		if len(d[0].Values()) != 1 {
			t.Error("failed to parse value")
			return
		}

		if v, ok := d[0].Values()[0].(string); !ok || v != "foo" {
			t.Error("failed to parse value")
			t.Log("value", v)
			t.Log("is string", ok)
		}
	})

	t.Run("empty definition", func(t *testing.T) {
		const code = `export "empty" = define()`

		buf := bytes.NewBufferString(code)
		d, err := Parse(&registry.Registry{}, buf)
		if err != nil {
			t.Error(err)
			return
		}

		if len(d) != 1 {
			t.Error("invalid number of definitions")
			return
		}

		if d[0].Path() != "empty" {
			t.Error("failed to parse path")
			return
		}

		if len(d[0].Values()) != 0 {
			t.Error("failed to parse value, length")
			return
		}
	})
}
