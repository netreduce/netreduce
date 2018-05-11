package engine

import (
	"errors"
	"sync"

	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/registry"
)

type Options struct{}

type Engine struct {
	mx       sync.Mutex
	protos   map[string]*proto
	registry *registry.Registry
	lru      *lru
}

var ErrNotFound = errors.New("not found")

func New(Options) *Engine {
	return &Engine{
		protos: make(map[string]*proto),
	}
}

func (e *Engine) checkEvict() {}

func (e *Engine) getExistingPlan(key string) (*plan, bool) {
	e.mx.Lock()
	defer e.mx.Unlock()

	proto, ok := e.protos[key]
	if !ok {
		return nil, false
	}

	e.lru.update(key)
	println("creating the existing plan")
	return proto.instance(), true
}

func (e *Engine) storeProto(key string, p *proto) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.checkEvict()
	e.protos[key] = p
	e.lru.update(key)
}

func (e *Engine) getPlan(c *Context) (*plan, bool) {
	key := c.definitionKey()
	p, ok := e.getExistingPlan(key)
	if ok {
		return p, true
	}

	println("no existing plan")
	d, ok := e.registry.Definition(key)
	if !ok {
		d, ok = c.definition()
		if !ok {
			return nil, false
		}
	}

	pp := createProto(d)
	e.storeProto(key, pp)
	println("creating new instance")
	return pp.instance(), true
}

func (e *Engine) Exec(c *Context) (data.Data, error) {
	p, ok := e.getPlan(c)
	if !ok {
		return data.Data{}, ErrNotFound
	}

	return p.exec(c)
}
