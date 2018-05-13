package url

import (
	"fmt"

	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/rules/eval"
)

const SetPathName = "setPath"

type (
	spspec struct{}
	sprule struct{
		arg interface{}
	}
)

func NewSetPath() rules.Spec {
	return spspec{}
}

func (s spspec) Name() string { return SetPathName }

func (s spspec) Instance(args []interface{}) (rules.Rule, error) {
	if len(args) != 1 || !eval.RuleOrString(args[0]) {
		return nil, fmt.Errorf("invalid arguments for setPath rule")
	}

	return sprule{arg: args[0]}, nil
}

func (r sprule) Exec(c rules.Context) (rules.Context, error) {
	p, err := eval.String(r.arg, c)
	if err != nil {
		return nil, err
	}

	return c.SetOutgoing(c.Outgoing().SetPath(p)), nil
}
