package registry

import (
	"fmt"

	"github.com/netreduce/netreduce/nred"
	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/rules/url"
	"github.com/netreduce/netreduce/rules/param"
)

type Registry struct{
	rules map[string]rules.Spec
}

func New() *Registry {
	return &Registry{
		rules: make(map[string]rules.Spec),
	}
}

func (r *Registry) Definition(key string) (nred.Definition, bool) { return nred.Definition{}, false }

func (r *Registry) Rule(name string) (rules.Spec, bool) {
	rule, ok := r.rules[name]
	return rule, ok
}

func (r *Registry) RegisterRule(rules ...rules.Spec) error {
	for _, ri := range rules {
		if _, ok := r.rules[ri.Name()]; ok {
			return fmt.Errorf("rule exists: %s", ri.Name())
		}

		r.rules[ri.Name()] = ri
	}

	return nil
}

func (r *Registry) RegisterBuiltinRules() error {
	return r.RegisterRule([]rules.Spec{
		url.NewSetURL(),
		url.NewSetPath(),
		url.NewAppendPath(),
		param.New(),
	}...)
}
