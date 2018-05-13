package engine

import (
	"github.com/netreduce/netreduce/nred"
	"github.com/netreduce/netreduce/registry"
)

type proto struct {
	definition nred.Definition
	registry *registry.Registry
}

func newProto(r *registry.Registry, d nred.Definition) *proto {
	return &proto{definition: d, registry: r}
}

func (p *proto) instance() *plan { return &plan{definition: p.definition, registry: p.registry} }
