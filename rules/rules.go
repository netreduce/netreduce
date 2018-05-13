package rules

import "github.com/netreduce/netreduce/data"

type Request interface {
	URL() string
	SetURL(string) Request
	Path() string
	SetPath(string) Request
}

type Context interface{
	Outgoing() Request // TODO: incoming/outgoing
	SetOutgoing(Request) Context
	HasResponse() bool
	SetResponse(data.Data) Context
	Response() data.Data
	HasValue() bool
	SetValue(interface{}) Context
	Value() interface{}
	Param(string) (interface{}, bool)
}

type Rule interface {
	Exec(Context) (Context, error)
}

type Spec interface {
	Name() string
	Instance(args []interface{}) (Rule, error)
}
