package nred

func rulesEq(r ...RuleSpec) bool {
	if len(r) < 2 {
		return true
	}

	if r[0].name != r[1].name {
		return false
	}

	if len(r[0].args) != len(r[1].args) {
		return false
	}

	for i := range r[0].args {
		if r0, ok := r[0].args[i].(RuleSpec); ok {
			if r1, ok := r[1].args[i].(RuleSpec); !ok || !rulesEq(r0, r1) {
				return false
			}
		} else if r[0].args[i] != r[1].args[i] {
			return false
		}
	}

	return rulesEq(r[1:]...)
}

func ruleSetsEq(left, right []RuleSpec) bool {
	if len(left) != len(right) {
		return false
	}

	rr := make([]RuleSpec, len(right))
	copy(rr, right)

	for len(left) > 0 {
		var found bool
		for i := len(rr) - 1; i >= 0; i-- {
			if !rulesEq(left[0], rr[i]) {
				continue
			}

			found = true
			rr = append(rr[:i], rr[i+1:]...)
			break
		}

		if !found {
			return false
		}

		left = left[1:]
	}

	return true
}

func queriesEq(left, right []QuerySpec) bool {
	if len(left) != len(right) {
		return false
	}

	rq := make([]QuerySpec, len(right))
	copy(rq, right)

	for len(left) > 0 {
		var found bool
		for i := len(rq) - 1; i >= 0; i-- {
			if !ruleSetsEq(left[0].rules, rq[i].rules) {
				continue
			}

			found = true
			rq = append(rq[:i], rq[i+1:]...)
			break
		}

		if !found {
			return false
		}

		left = left[1:]
	}

	return true
}

func fieldsEq(f ...Field) bool {
	if len(f) < 2 {
		return true
	}

	if f[0].typ != f[1].typ {
		return false
	}

	if f[0].name != f[1].name {
		return false
	}

	if d0, ok := f[0].value.(Definition); ok {
		if d1, ok := f[1].value.(Definition); !ok || !Eq(d0, d1) {
			return false
		}
	} else if f[0].value != f[1].value {
		return false
	}

	return fieldsEq(f[1:]...)
}

func fieldSetsEq(left, right []Field) bool {
	if len(left) != len(right) {
		return false
	}

	rf := make([]Field, len(right))
	copy(rf, right)

	for len(left) > 0 {
		var found bool
		for i := len(rf) - 1; i >= 0; i-- {
			if !fieldsEq(left[0], rf[i]) {
				continue
			}

			found = true
			rf = append(rf[:i], rf[i+1:]...)
			break
		}

		if !found {
			return false
		}

		left = left[1:]
	}

	return true
}

func Eq(d ...Definition) bool {
	if len(d) < 2 {
		return true
	}

	if d[0].name != d[1].name {
		return false
	}

	if d[0].value != d[1].value {
		return false
	}

	if !queriesEq(d[0].queries, d[1].queries) {
		return false
	}

	if !fieldSetsEq(d[0].fields, d[1].fields) {
		return false
	}

	if !ruleSetsEq(d[0].rules, d[1].rules) {
		return false
	}

	return Eq(d[1:]...)
}
