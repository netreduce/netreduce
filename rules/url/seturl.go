package url

import (
	"fmt"

	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/rules/eval"
)

const SetURLName = "setURL"

type (
	suspec struct {}
	surule struct {
		arg interface{}
	}
)

func NewSetURL() rules.Spec {
	return suspec{}
}

func (s suspec) Name() string { return SetURLName }

func (s suspec) Instance(args []interface{}) (rules.Rule, error) {
	if len(args) != 1 || !eval.RuleOrString(args[0]) {
		return nil, fmt.Errorf("invalid arguments for setURL rule")
	}

	return surule{arg: args[0]}, nil
}

func (r surule) Exec(c rules.Context) (rules.Context, error) {
	u, err := eval.String(r.arg, c)
	if err != nil {
		return nil, err
	}

	return c.SetOutgoing(c.Outgoing().SetURL(u)), nil
}
