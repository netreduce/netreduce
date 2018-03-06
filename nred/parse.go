package nred

import (
	"io"
	"errors"
	"strconv"
	"bytes"

	"github.com/netreduce/netreduce/nred/parser"
	"github.com/netreduce/netreduce/registry"
)

var (
	errNotImplemented = errors.New("not implemented")
	errInvalidExpression = errors.New("invalid expression")
	errDuplicateDefinition = errors.New("duplicate definition")
	errDuplicateExport = errors.New("duplicate export")
)

func isReserved(r *registry.Registry, name string) bool {
	return false
}

func parseInt(n *parser.Node) (int, error) {
	return strconv.Atoi(n.Text())
}

func parseFloat(n *parser.Node) (float64, error) {
	return strconv.ParseFloat(n.Text(), 64)
}

func unescapeString(s string) string {
	var (
		us []byte
		escaped bool
	)

	for _, b := range []byte(s) {
		if escaped {
			switch b {
			case 'b':
				us = append(us, '\b')
			case 'f':
				us = append(us, '\f')
			case 'n':
				us = append(us, '\n')
			case 'r':
				us = append(us, '\r')
			case 't':
				us = append(us, '\t')
			case 'v':
				us = append(us, '\v')
			default:
				us = append(us, b)
			}

			escaped = false
			continue
		}

		if b == '\\' {
			escaped = true
			continue
		}

		us = append(us, b)
	}

	return string(us)
}

func parseString(n *parser.Node) string {
	text := n.Text()
	return unescapeString(text[1:len(text) - 1])
}

func parseCompositeExpression(n *parser.Node) (interface{}, error) {
	switch n.Nodes[0].Text() {
	case "define":
		return Define(), nil
	default:
		// constant, field, contains, link, query, mapping, or symbol representing a reference
		return nil, errNotImplemented
	}
}

func parseExpression(n *parser.Node) (interface{}, error) {
	switch n.Name {
	case "int":
		i, err := parseInt(n)
		if err != nil {
			return nil, err
		}

		return Define(i), nil
	case "float":
		f, err := parseFloat(n)
		if err != nil {
			return nil, err
		}

		return Define(f), nil
	case "string":
		s := parseString(n)
		return Define(s), nil
	case "composite-expression":
		return parseCompositeExpression(n)
	default:
		return nil, errNotImplemented
	}
}

func parseLocal(n *parser.Node) (string, interface{}, error) {
	return "", nil, errNotImplemented
}

func parseExport(n *parser.Node) (d Definition, err error) {
	path := parseString(n.Nodes[0])

	var expression interface{}
	if expression, err = parseExpression(n.Nodes[1]); err != nil {
		return
	}

	var ok bool
	if d, ok = expression.(Definition); !ok {
		err = errInvalidExpression
	}

	d = Export(path, d)
	return
}

func substituteLocal(exported map[string]Definition, local map[string]interface{}) ([]Definition, error) {
	var defs []Definition
	for _, d := range exported {
		defs = append(defs, d)
	}

	return defs, nil
}

func Parse(reg *registry.Registry, r io.Reader) ([]Definition, error) {
	n, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	local := make(map[string]interface{})
	exported := make(map[string]Definition)
	for _, ni := range n.Nodes {
		switch ni.Name {
		case "local":
			name, expression, err := parseLocal(ni)
			if err != nil {
				return nil, err
			}

			if _, ok := local[name]; ok || isReserved(reg, name) {
				return nil, errDuplicateDefinition
			}

			local[name] = expression
		case "export":
			definition, err := parseExport(ni)
			if err != nil {
				return nil, err
			}

			path := definition.Path()
			if _, ok := exported[path]; ok {
				return nil, errDuplicateExport
			}

			exported[path] = definition
		default:
			return nil, errors.New("unsupported expression")
		}
	}

	defs, err := substituteLocal(exported, local)
	if err != nil {
		return nil, err
	}

	for i := range defs {
		defs[i] = Normalize(defs[i])
	}

	return defs, nil
}

func ParseString(reg *registry.Registry, s string) ([]Definition, error) {
	return Parse(reg, bytes.NewBufferString(s))
}
