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
	definition define.Definition
	childProtos map[string]*node
	query         *query

	result        data.Struct
	backendFields []define.Field
	children      map[string]*node
}

type plan struct {
	queries []*query
	root    *node
}

var errCanceled = errors.New("canceled")

func newQueryProto(q define.QuerySpec, r *Registry) (*query, error) {
	conn, ok := q.Connector().(Connector)
	if !ok {
		return nil, errors.New("invalid query: no valid connector defined")
	}

	client, ok := r.clients[conn]
	if !ok {
		return nil, errors.New("invalid query: unregistered connector")
	}

	ctx := ConnectorContext{Client: client}

	return &query{
		connector:        conn,
		connectorContext: ctx,
	}, nil
}

func (q *query) instance() *query {
	i := *q
	i.buffer = make(chan queryResult, 1)
	return &i
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

func newNodeProto(d define.Definition, r *Registry) (*node, error) {
	n := &node{
		definition: d,
		childProtos: make(map[string]*node),
	}

	for _, f := range d.Fields() {
		switch f.Type() {
		case define.ConstantField:
		case define.OneChildField:

			// TODO:
			// - support multi def fields
			// - support fields (and responses) without definitions
			fieldDefs := f.Definitions()
			if len(fieldDefs) != 1 {
				return nil, errNotImplemented
			}

			child, err := newNodeProto(fieldDefs[0], r)
			if err != nil {
				return nil, err
			}

			n.childProtos[f.Name()] = child

		default:
			switch f.Type() {
			case define.IntField, define.StringField:
				n.backendFields = append(n.backendFields, f)
			default:
				return nil, errors.New("not implemented field type")
			}
		}
	}

	query := d.Query()
	if query == define.ZeroQuery && len(n.backendFields) > 0 {
		return nil, errors.New("invalid definition: missing query")
	}

	if query != define.ZeroQuery {
		q, err := newQueryProto(query, r)
		if err != nil {
			return nil, err
		}

		n.query = q
	}

	return n, nil
}

func (n *node) instance() *node {
	i := &node{
		result:   make(data.Struct),
		children: make(map[string]*node),
	}

	for _, f := range n.definition.Fields() {
		switch f.Type() {
		case define.ConstantField:
			i.result[f.Name()] = f.Value()
		case define.OneChildField:
			i.children[f.Name()] = n.childProtos[f.Name()].instance()
		default:
			i.backendFields = append(i.backendFields, f)
		}
	}

	if n.query != nil {
		i.query = n.query.instance()
	}

	return i
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
		}

		if err != nil {
			return nil, err
		}

		n.result[f.Name()] = v
	}

	return n.result, nil
}

func newProto(root define.Definition, r *Registry) (*plan, error) {
	rootNode, err := newNodeProto(root, r)
	if err != nil {
		return nil, err
	}

	return &plan{root: rootNode}, nil
}

func (p *plan) instance() *plan {
	root := p.root.instance()

	// TODO: remove duplicates
	queries := root.queries()

	return &plan{
		root: root,
		queries: queries,
	}
}

func (p *plan) execute(ctx RequestContext) (interface{}, error) {
	for i := range p.queries {
		go p.queries[i].execute(ctx)
	}

	return p.root.construct(ctx)
}
