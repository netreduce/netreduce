package nred

import (
	"errors"
	"fmt"
)

const (
	letSymbol      = "let"
	exportSymbol   = "export"
	defineSymbol   = "define"
	querySymbol    = "query"
	containsSymbol = "contains"
	constSymbol    = "const"
	intSymbol      = "int"
	floatSymbol    = "float"
	numberSymbol   = "number"
	stringSymbol   = "string"
	boolSymbol     = "bool" // TODO
	fieldSymbol    = "field"
	trueSymbol     = "true"
	falseSymbol    = "false"
	nilSymbol      = "nil"
)

const urlConnectorName = "url"

var (
	errFieldArgsCount = errors.New("invalid number of field arguments")
	errFieldName      = errors.New("invalid field name argument")
)

func Reserved(name string) bool {
	for _, r := range []string{
		letSymbol,
		exportSymbol,
		defineSymbol,
		querySymbol,
		containsSymbol,
		constSymbol,
		intSymbol,
		floatSymbol,
		numberSymbol,
		stringSymbol,
		boolSymbol, // TODO
		fieldSymbol,
		trueSymbol,
		falseSymbol,
		nilSymbol,
	} {
		if name == r {
			return true
		}
	}

	return false
}

func getFieldName(exp expression) (string, error) {
	if exp.typ != stringExp {
		return "", errFieldName
	}

	return exp.primitive.(string), nil
}

func getFieldNameOnly(args []expression) (name string, err error) {
	if len(args) != 1 {
		err = errFieldArgsCount
		return
	}

	name, err = getFieldName(args[0])
	return
}

func defineConst(args []expression) (d Definition, err error) {
	if len(args) != 2 {
		err = errFieldArgsCount
		return
	}

	name, err := getFieldName(args[0])
	if err != nil {
		return
	}

	switch args[1].typ {
	case intExp, floatExp, opaqueNumberExp, stringExp, trueExp, falseExp, nilExp:
		d = Define(Const(name, args[1].primitive))
	default:
		err = errors.New("invalid const field value")
	}

	return
}

func defineNamedField(args []expression, field func(string) Field) (d Definition, err error) {
	var name string
	if name, err = getFieldNameOnly(args); err != nil {
		return
	}

	d = Define(field(name))
	return
}

func untypeDefs(d []Definition) []interface{} {
	var i []interface{}
	for _, di := range d {
		i = append(i, di)
	}

	return i
}

func defineContains(args []expression) (d Definition, err error) {
	if len(args) == 0 {
		err = errFieldArgsCount
		return
	}

	name, err := getFieldName(args[0])
	if err != nil {
		return
	}

	defs, err := defineAll(args[1:])
	if err != nil {
		return
	}

	d = Define(Contains(name, Define(untypeDefs(defs)...)))
	return
}

func defineField(typ string, args []expression) (Definition, error) {
	switch typ {
	case constSymbol:
		return defineConst(args)
	case fieldSymbol:
		return defineNamedField(args, Generic)
	case intSymbol:
		return defineNamedField(args, Int)
	case floatSymbol:
		return defineNamedField(args, Float)
	case numberSymbol:
		return defineNamedField(args, Number)
	case stringSymbol:
		return defineNamedField(args, String)
	default:
		return defineContains(args)
	}
}

func defineRuleArgs(args []expression) ([]interface{}, error) {
	var argValues []interface{}
	for _, argExp := range args {
		switch argExp.typ {
		case symbolExp, compositeExp:
			r, err := defineRule(argExp)
			if err != nil {
				return nil, err
			}

			argValues = append(argValues, r)
		default:
			argValues = append(argValues, argExp.primitive)
		}
	}

	return argValues, nil
}

func defineCompositeRule(exp expression) (r RuleSpec, err error) {
	switch exp.children[0].typ {
	case symbolExp:
		name := exp.children[0].primitive.(string)
		var args []interface{}
		args, err = defineRuleArgs(exp.children[1:])
		if err != nil {
			return
		}

		r = Rule(name, args...)
	case compositeExp:
		err = errors.New("currying of rules not allowed")
	default:
		err = fmt.Errorf("invalid rule: %v", exp.children[0].primitive)
	}

	return
}

func defineRule(exp expression) (r RuleSpec, err error) {
	switch exp.typ {
	case compositeExp:
		return defineCompositeRule(exp)
	default:
		// symbolExp

		name := exp.primitive.(string)
		if Reserved(name) {
			err = fmt.Errorf("reserved word: %s", name)
			return
		}

		r = Rule(name)
	}

	return
}

func defineQueryRule(exp expression) (RuleSpec, error) {
	switch exp.typ {
	case symbolExp, compositeExp:
		return defineRule(exp)
	default:
		return Rule(urlConnectorName, exp.primitive), nil
	}
}

func defineQuery(args []expression) (d Definition, err error) {
	var r []RuleSpec
	for _, arg := range args {
		var rule RuleSpec
		rule, err = defineQueryRule(arg)
		if err != nil {
			return
		}

		r = append(r, rule)
	}

	d = Define(Query(r...))
	return
}

func defineBySymbol(name string, args []expression) (d Definition, err error) {
	switch name {
	case constSymbol, fieldSymbol, intSymbol, floatSymbol, numberSymbol, stringSymbol, containsSymbol:
		d, err = defineField(name, args)
	case querySymbol:
		d, err = defineQuery(args)
	case defineSymbol:
		var ds []Definition
		ds, err = defineAll(args)
		if err != nil {
			return
		}

		d = Define(untypeDefs(ds)...)
	default:
		var a []interface{}
		if a, err = defineRuleArgs(args); err != nil {
			return
		}

		d = Define(Rule(name, a...))
	}

	return
}

func defineComposite(exp expression) (d Definition, err error) {
	switch exp.children[0].typ {
	case symbolExp:
		d, err = defineBySymbol(exp.children[0].primitive.(string), exp.children[1:])
	default:
		d, err = define(exp.children[0])
		if err != nil {
			return
		}

		var args []Definition
		args, err = defineAll(exp.children[1:])
		if err != nil {
			return
		}

		d = Define(append([]interface{}{d}, untypeDefs(args)...)...)
	}

	return
}

func define(exp expression) (d Definition, err error) {
	switch exp.typ {
	case opaqueNumberExp:
		d = Define(OpaqueNumber(exp.primitive.(string)))
	case trueExp:
		d = Define(true)
	case falseExp:
		d = Define(false)
	case nilExp:
		d = Define(NilValue)
	case symbolExp:
		d, err = defineBySymbol(exp.primitive.(string), nil)
	case compositeExp:
		d, err = defineComposite(exp)
	default:
		// int, float, string
		d = Define(exp.primitive)
	}

	return
}

func defineAll(exps []expression) (d []Definition, err error) {
	var di Definition
	for _, exp := range exps {
		if di, err = define(exp); err != nil {
			return
		}

		d = append(d, di)
	}

	return
}

func defineExports(exports namedExpressions) ([]Definition, error) {
	var d []Definition
	for name, exp := range exports {
		di, err := define(exp)
		if err != nil {
			return nil, err
		}

		di = Export(name, di)
		d = append(d, di)
	}

	return d, nil
}
