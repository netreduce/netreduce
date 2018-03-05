package netreduce

import (
	"errors"

	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/define"
)

type queryResult struct {
	data interface{}
	err  error
}

type query struct {
	connector        Connector
	connectorContext ConnectorContext
	buffer           chan queryResult
}

type node struct {
	result        data.Struct
	backendFields []define.Field
	children      map[string]*node
	query         *query
}

type plan struct {
	queries []*query
	root    *node
}

var errCanceled = errors.New("canceled")

func newQuery(q define.QuerySpec, r *Registry) *query {
	conn, ok := q.Connector().(Connector)
	if !ok {
		// TODO: validate when separating phases
		panic(errors.New("invalid query: no valid connector defined"))
	}

	client, ok := r.clients[conn]
	if !ok {
		// TODO: validate when separating phases
		panic(errors.New("invalid query: unregistered connector"))
	}

	ctx := ConnectorContext{Client: client}

	return &query{
		buffer:           make(chan queryResult, 1),
		connector:        conn,
		connectorContext: ctx,
	}
}

func (q *query) execute(ctx RequestContext) {
	d, err := q.connector.Call(q.connectorContext, ctx)
	q.buffer <- queryResult{data: d, err: err}
}

func (q *query) result(ctx RequestContext) queryResult {
	select {
	case r := <-q.buffer:
		q.buffer <- r
		return r
	case <-ctx.Canceled():
		return queryResult{err: errCanceled}
	}
}

func newNode(d define.Definition, r *Registry) *node {
	n := &node{
		result:   make(data.Struct),
		children: make(map[string]*node),
	}

	for _, f := range d.Fields() {
		switch f.Type() {
		case define.ConstantField:
			n.result[f.Name()] = f.Value()
		case define.OneChildField:

			// TODO:
			// - validate when separating phases
			// - support multi def fields
			// - support fields (and responses) without definitions
			fieldDefs := f.Definitions()
			if len(fieldDefs) != 1 {
				panic(errNotImplemented)
			}

			n.children[f.Name()] = newNode(fieldDefs[0], r)
		default:
			n.backendFields = append(n.backendFields, f)
		}
	}

	query := d.Query()
	if query == define.ZeroQuery && len(n.backendFields) > 0 {
		// TODO: validate when separating phases
		panic(errors.New("invalid definition: missing query"))
	}

	if query != define.ZeroQuery {
		n.query = newQuery(query, r)
	}

	return n
}

func (n *node) queries() []*query {
	var q []*query

	if n.query != nil {
		q = append(q, n.query)
	}

	for _, c := range n.children {
		q = append(q, c.queries()...)
	}

	return q
}

func (n *node) construct(ctx RequestContext) (interface{}, error) {
	for field, child := range n.children {
		d, err := child.construct(ctx)
		if err != nil {
			return nil, err
		}

		n.result[field] = d
	}

	if n.query == nil {
		return n.result, nil
	}

	r := n.query.result(ctx)

	// TODO: optional queries
	if r.err != nil {
		if r.err != errCanceled {
			// TODO:
			// - cleanup canceling
			// - preserve the first original error
			ctx.Cancel()
		}

		return nil, r.err
	}

	if id, err := data.GetString(r.data, "id"); err == nil {
		n.result["id"] = id
	}

	for _, f := range n.backendFields {
		var (
			v   interface{}
			err error
		)

		switch f.Type() {
		case define.IntField:
			v, err = data.GetInt(r.data, f.Name())
		case define.StringField:
			v, err = data.GetString(r.data, f.Name())
		default:
			err = errors.New("not implemented field type")
		}

		if err != nil {
			return nil, err
		}

		n.result[f.Name()] = v
	}

	return n.result, nil
}

func newPlan(root define.Definition, r *Registry) *plan {
	p := &plan{}
	p.root = newNode(root, r)

	// TODO: remove duplicates
	p.queries = p.root.queries()

	return p
}

func (p *plan) execute(ctx RequestContext) (interface{}, error) {
	c := <-ctx.cancel
	ctx.cancel <- c
	for i := range p.queries {
		go p.queries[i].execute(ctx)
	}

	return p.root.construct(ctx)
}
