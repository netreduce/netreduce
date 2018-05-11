package nred

import (
	"fmt"
	"io"
	"strconv"

	"github.com/netreduce/netreduce/nred/parser"
)

type expressionType int

const (
	intExp expressionType = iota
	floatExp
	opaqueNumberExp
	stringExp
	trueExp
	falseExp
	nilExp
	symbolExp
	compositeExp
)

type expression struct {
	typ       expressionType
	primitive interface{}
	children  []expression
}

type namedExpressions map[string]expression

func primitive(typ expressionType, value interface{}) expression {
	return expression{
		typ:       typ,
		primitive: value,
	}
}

func composite(children []expression) expression {
	return expression{
		typ:      compositeExp,
		children: children,
	}
}

func parsePrimitive(n *parser.Node) (exp expression, err error) {
	typ := n.Name
	switch typ {
	case "int":
		exp.typ = intExp
		var v int
		v, err = strconv.Atoi(n.Text())
		if err != nil {
			return
		}

		exp.primitive = v
	case "float":
		exp.typ = floatExp
		var v float64
		v, err = strconv.ParseFloat(n.Text(), 64)
		if err != nil {
			return
		}

		exp.primitive = v
	case "string":
		exp.typ = stringExp
		t := n.Text()
		exp.primitive = unescapeString(t[1 : len(t)-1])
	case "symbol":
		switch n.Text() {
		case trueSymbol:
			exp.typ = trueExp
		case falseSymbol:
			exp.typ = falseExp
		case nilSymbol:
			exp.typ = nilExp
		default:
			exp.typ = symbolExp
			exp.primitive = n.Text()
		}
	}

	return
}

func parseOpaqueNumber(n *parser.Node) (exp expression, ok bool) {
	if len(n.Nodes) != 2 {
		return
	}

	switch n.Nodes[1].Name {
	case "int", "float":
		exp.typ = opaqueNumberExp
		exp.primitive = n.Nodes[1].Text()
		ok = true
	}

	return
}

func parseComposite(n *parser.Node) (exp expression, err error) {
	if n.Nodes[0].Text() == numberSymbol {
		var ok bool
		exp, ok = parseOpaqueNumber(n)
		if ok {
			return
		}
	}

	exp.typ = compositeExp
	for i := range n.Nodes {
		var ce expression
		ce, err = parseExpression(n.Nodes[i])
		if err != nil {
			return
		}

		exp.children = append(exp.children, ce)
	}

	return
}

func parseExpression(n *parser.Node) (expression, error) {
	if n.Name == "composite-expression" {
		return parseComposite(n)
	}

	return parsePrimitive(n)
}

func parseNodes(n []*parser.Node) (namedExpressions, namedExpressions, error) {
	local := make(namedExpressions)
	export := make(namedExpressions)

	for i := range n {
		p, err := parsePrimitive(n[i].Nodes[0])
		if err != nil {
			return nil, nil, err
		}

		name := p.primitive.(string)
		value, err := parseExpression(n[i].Nodes[1])
		if err != nil {
			return nil, nil, err
		}

		switch n[i].Name {
		case "local":
			if _, has := local[name]; has {
				return nil, nil, fmt.Errorf("duplicate local definition: %s", name)
			}

			local[name] = value
		case "export":
			if _, has := export[name]; has {
				return nil, nil, fmt.Errorf("duplicate export: %s", name)
			}

			export[name] = value
		}
	}

	return local, export, nil
}

func Parse(r io.Reader) ([]Definition, error) {
	n, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	if len(n.Nodes) == 1 && n.Nodes[0].Name == "single-expression" {
		e, err := parseExpression(n.Nodes[0])
		if err != nil {
			return nil, err
		}

		d, err := define(e)
		if err != nil {
			return nil, err
		}

		return []Definition{d}, nil
	}

	local, exports, err := parseNodes(n.Nodes)
	if err != nil {
		return nil, err
	}

	if err = resolve(local, exports); err != nil {
		return nil, err
	}

	return defineExports(exports)
}
