package param

import (
	"fmt"

	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/rules/eval"
)

const Name = "param"

type (
	spec struct{}
	rule struct{
		arg interface{}
	}
)

func New() rules.Spec { return spec{} }

func (s spec) Name() string { return Name }

func (s spec) Instance(args []interface{}) (rules.Rule, error) {
	if len(args) != 1 || !eval.RuleOrString(args[0]) {
		return nil, fmt.Errorf("invalid arguments for param rule")
	}

	return rule{arg: args[0]}, nil
}

func (r rule) Exec(c rules.Context) (rules.Context, error) {
	name, err := eval.String(r.arg, c)
	if err != nil {
		return nil, err
	}

	v, ok := c.Param(name)
	if !ok {
		return nil, fmt.Errorf("param not found: %s", name)
	}

	return c.SetValue(v), nil
}
