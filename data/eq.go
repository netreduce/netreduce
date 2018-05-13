package data

import "bytes"

func fieldsEq(f0, f1 map[string]Data) bool {
	for n, v0 := range f0 {
		if v1, ok := f1[n]; !ok || !Eq(v0, v1) {
			return false
		}
	}

	for n := range f1 {
		if _, ok := f0[n]; !ok {
			return false
		}
	}

	return true
}

func itemsEq(i0, i1 []Data) bool {
	if len(i0) != len(i1) {
		return false
	}

	for i, item := range i0 {
		if !Eq(item, i1[i]) {
			return false
		}
	}

	return true
}

func Eq(d ...Data) bool {
	if len(d) < 2 {
		return true
	}

	if d[0].typ != d[1].typ {
		return false
	}

	if d[0].primitive != d[1].primitive {
		return false
	}

	if !fieldsEq(d[0].fields, d[1].fields) {
		return false
	}

	if !itemsEq(d[0].items, d[1].items) {
		return false
	}

	if !bytes.Equal(d[0].json, d[1].json) {
		return false
	}

	return Eq(d[1:]...)
}
