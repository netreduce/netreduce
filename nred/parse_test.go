package nred

import (
	"testing"
	"bytes"
)

func TestParse(t *testing.T) {
	const code = `export "/foo" "foo"`

	buf := bytes.NewBufferString(code)
	_, err := parse(buf)
	if err != nil {
		t.Error(err)
	}
}
