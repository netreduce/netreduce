package eval

import (
	"fmt"

	"github.com/netreduce/netreduce/rules"
)

func RuleOrString(v interface{}) bool {
	switch v.(type) {
	case rules.Rule, string:
		return true
	default:
		return false
	}
}

func String(v interface{}, c rules.Context) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}

	r, ok := v.(rules.Rule)
	if !ok {
		return "", fmt.Errorf("unexpected value: %v", v)
	}

	c, err := r.Exec(c)
	if err != nil {
		return "", err
	}

	if !c.HasValue() {
		return "", fmt.Errorf("unexpected value: %v", v)
	}

	v = c.Value()
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("unexpected value: %v", v)
	}

	return s, nil
}
