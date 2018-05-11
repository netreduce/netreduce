package engine

import (
	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/nred"
)

type proto struct {
	data data.Data
}

func createProto(d nred.Definition) *proto {
	return &proto{data: d.GetValue()}
}

func (p *proto) instance() *plan { return &plan{data: p.data} }
