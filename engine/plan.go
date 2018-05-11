package engine

import "github.com/netreduce/netreduce/data"

type plan struct {
	data data.Data
}

func (p *plan) exec(*Context) (data.Data, error) { return p.data, nil }
