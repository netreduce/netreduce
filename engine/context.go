package engine

import "github.com/netreduce/netreduce/nred"

type Context struct {
	def    nred.Definition
	defSet bool
	params map[string][]string
}

func FromQuery(d nred.Definition, params map[string][]string) *Context {
	return &Context{def: d, defSet: true, params: params}
}

func (c *Context) definitionKey() string { return "" }

func (c *Context) definition() (nred.Definition, bool) {
	return c.def, c.defSet
}
