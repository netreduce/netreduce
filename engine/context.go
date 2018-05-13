package engine

import (
	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/rules"
)

type context struct {
	incoming request
	outgoing rules.Request
	params map[string]interface{}
	hasValue bool
	value interface{}
}

func newContext(incoming request, params map[string]interface{}) context {
	return context{
		incoming: incoming,
		outgoing: request{},
		params:  params,
	}
}

func (c context) Outgoing() rules.Request { return c.outgoing }

func (c context) SetOutgoing(r rules.Request) rules.Context {
	c.outgoing = r
	return c
}

func (c context) HasResponse() bool { return false }
func (c context) SetResponse(data.Data) rules.Context { return c }
func (c context) Response() data.Data { return data.Zero() }
func (c context) HasValue() bool { return c.hasValue }

func (c context) SetValue(v interface{}) rules.Context {
	c.value = v
	c.hasValue = true
	return c
}

func (c context) Value() interface{} { return c.value }

func (c context) Param(name string) (interface{}, bool) {
	v, ok := c.params[name]
	return v, ok
}
