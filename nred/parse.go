package nred

import (
	"io"

	"github.com/netreduce/netreduce/nred/parser"
)

var errNotImplemented = errors.New("not implemented")

func parseLocal(n *parser.Node) (string, interface{}, error) {
	return "", nil, errNotImplemented
}

// substituted
func Parse(r io.Reader) ([]Definition, error) {
	n, err := parser.Parse(r)
	if err != nil {
		return err
	}

	local := make(map[string]interface{})
	exported := make(map[string]Definition)
	for i, ni := range n.Nodes {
		switch ni.Name {
		case "local":
			name, expression, err := parseLocal(ni)
			if err != nil {
				return nil, err
			}
		case "export":
			path, definition, err := parseExport(ni)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unsupported expression")
		}
	}

	return substitueLocal(exported, local)
}

func ParseString(s string) ([]Definition, error) {
	return nil, nil
}
