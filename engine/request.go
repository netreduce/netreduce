package engine

import (
	"io"

	"github.com/netreduce/netreduce/rules"
)

type (
	request struct {
		url string
	}

	Incoming struct {
		request
		params map[string]interface{}
	}
)

func NewIncoming(method, path string, headers map[string][]string, body io.Reader) Incoming {
	return Incoming{}
}

func (r request) URL() string { return r.url }

func (r request) SetURL(u string) rules.Request {
	r.url = u
	return r
}

func (r request) Path() string { return "" }
func (r request) SetPath(string) rules.Request { return r }
