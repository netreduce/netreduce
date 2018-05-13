package url

import (
	"fmt"
	"path"

	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/rules/eval"
)

const AppendPathName = "appendPath"

type (
	apspec struct{}
	aprule struct{
		arg interface{}
	}
)

func NewAppendPath() rules.Spec {
	return apspec{}
}

func (s apspec) Name() string { return AppendPathName }

func (s apspec) Instance(args []interface{}) (rules.Rule, error) {
	if len(args) != 1 || !eval.RuleOrString(args[0]) {
		return nil, fmt.Errorf("invalid arguments for appendPath rule")
	}

	return aprule{arg: args[0]}, nil
}

func (r aprule) Exec(c rules.Context) (rules.Context, error) {
	p, err := eval.String(r.arg, c)
	if err != nil {
		return nil, err
	}

	return c.SetOutgoing(c.Outgoing().SetPath(path.Join(c.Outgoing().Path(), p))), nil
}
