package engine

import (
	"errors"
	"sync"

	"github.com/netreduce/netreduce/nred"
	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/registry"
)

type Options struct{
	registry *registry.Registry
}

type Engine struct {
	mx       sync.Mutex
	protos   map[string]*proto
	registry *registry.Registry
	lru      *lru
}

var ErrNotFound = errors.New("not found")

func New(o Options) *Engine {
	return &Engine{
		registry: o.registry,
		protos: make(map[string]*proto),
	}
}

func (e *Engine) checkEvict() {}

func (e *Engine) getProto(key string) (*proto, bool) {
	e.mx.Lock()
	defer e.mx.Unlock()

	proto, ok := e.protos[key]
	if !ok {
		return nil, false
	}

	e.lru.update(key)
	return proto, true
}

func (e *Engine) storeProto(key string, p *proto) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.checkEvict()
	e.protos[key] = p
	e.lru.update(key)
}

func (e *Engine) Exec(key string, req Incoming) (data.Data, error) {
	c := newContext(req.request, req.params)
	p, ok := e.getProto(key)
	if ok {
		return p.instance().exec(c)
	}

	d, ok := e.registry.Definition(key)
	if !ok {
		return data.Zero(), ErrNotFound
	}

	p = newProto(e.registry, d)
	e.storeProto(key, p)
	return p.instance().exec(c)
}

// TODO:
// - validation
// - IncomingRequest
func (e *Engine) ExecDefinition(d nred.Definition, req Incoming) (data.Data, error) {
	c := newContext(req.request, req.params)
	key := d.String()
	p, ok := e.getProto(key)
	if ok {
		return p.instance().exec(c)
	}

	p = newProto(e.registry, d)
	e.storeProto(key, p)
	return p.instance().exec(c)
}
