package engine

import (
	"fmt"

	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/nred"
	"github.com/netreduce/netreduce/registry"
	"github.com/netreduce/netreduce/rules"
	"github.com/netreduce/netreduce/nrhttp"
)

type plan struct {
	definition nred.Definition
	registry *registry.Registry
}

func resolveRule(r *registry.Registry, def nred.Rule) (rules.Rule, error) {
	rspec, ok := r.Rule(def.Name())
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", def.Name())
	}

	var args []interface{}
	for _, a := range def.Args() {
		adef, ok := a.(nred.Rule)
		if !ok {
			args = append(args, a)
			continue
		}

		ar, err := resolveRule(r, adef)
		if err != nil {
			return nil, err
		}

		args = append(args, ar)
	}

	return rspec.Instance(args)
}

func (p *plan) exec(ec context) (d data.Data, err error) {
	var c rules.Context = ec
	d = p.definition.GetValue()
	if !data.IsZero(d) {
		return
	}

	d = data.Zero()

	for _, f := range p.definition.Fields() {
		if f.Type() == nred.ConstField {
			if d, err = d.SetField(f.Name(), f.Value()); err != nil {
				return
			}
		}
	}

	for _, q := range p.definition.Queries() {
		for _, rdef := range q.Rules() {
			var r rules.Rule
			if r, err = resolveRule(p.registry, rdef); err != nil {
				return
			}

			if c, err = r.Exec(c); err != nil {
				return
			}
		}

		var dq data.Data
		if c.HasResponse() {
			dq = c.Response()
		} else {
			dq, err = nrhttp.Get(c.Outgoing().URL())
			if err != nil {
				return
			}
		}

		if d, err = d.Merge(dq); err != nil {
			return
		}
	}

	return
}
