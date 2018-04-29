package nred

import "fmt"

func resolveSymbol(
	all namedExpressions,
	resolving map[string]bool,
	exp expression,
) (expression, error) {
	name := exp.primitive.(string)
	if _, has := resolving[name]; has {
		return exp, fmt.Errorf("circular dependency: %s", name)
	}

	resolving[name] = true
	defer func() {
		delete(resolving, name)
	}()

	value, ok := all[name]
	if !ok {
		// not a local symbol, checked in a separate pass
		return exp, nil
	}

	return resolveExpression(all, resolving, value)
}

func resolveComposite(
	all namedExpressions,
	resolving map[string]bool,
	exp expression,
) (expression, error) {
	for i := range exp.children {
		cexp, err := resolveExpression(all, resolving, exp.children[i])
		if err != nil {
			return exp, err
		}

		exp.children[i] = cexp
	}

	return exp, nil
}

func resolveExpression(
	all namedExpressions,
	resolving map[string]bool,
	exp expression,
) (expression, error) {
	switch exp.typ {
	case symbolExp:
		return resolveSymbol(all, resolving, exp)
	case compositeExp:
		return resolveComposite(all, resolving, exp)
	default:
		return exp, nil
	}
}

func resolveLocal(local namedExpressions) error {
	resolving := make(map[string]bool)
	for name := range local {
		if Reserved(name) {
			return fmt.Errorf("reserved word: %s", name)
		}

		exp, err := resolveExpression(local, resolving, local[name])
		if err != nil {
			return err
		}

		local[name] = exp
	}

	return nil
}

func resolveExports(local, exports namedExpressions) error {
	resolving := make(map[string]bool)
	for name := range exports {
		exp, err := resolveExpression(local, resolving, exports[name])
		if err != nil {
			return err
		}

		exports[name] = exp
	}

	return nil
}

func resolve(local, exports namedExpressions) error {
	if err := resolveLocal(local); err != nil {
		return err
	}

	return resolveExports(local, exports)
}
