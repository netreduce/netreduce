package nred

import (
	"bytes"
	"testing"
)

func TestParse(t *testing.T) {
	const code = `export "/foo" "foo"`

	buf := bytes.NewBufferString(code)
	_, err := Parse(buf)
	if err != nil {
		t.Error(err)
	}
}
