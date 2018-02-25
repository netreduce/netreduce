package data

import "fmt"

type Struct map[string]interface{}

type List []interface{}

func Int(data interface{}, name string) int {
	i, err := GetInt(data, name)
	if err != nil {
		panic(err)
	}

	return i
}

func GetInt(data interface{}, name string) (int, error) {
	var (
		v interface{}
		ok bool
	)

	switch dt := data.(type) {
	case map[string]interface{}:
		v, ok = dt[name]
	case Struct:
		v, ok = dt[name]
	default:
		return 0, fmt.Errorf("not a struct")
	}

	missingOrInvalid := func() error {
		return fmt.Errorf("missing or invalid field: %s", name)
	}

	if !ok {
		return 0, missingOrInvalid()
	}

	vi, ok := v.(int)
	if !ok {
		vf, ok := v.(float64)
		if !ok {
			return 0, missingOrInvalid()
		}

		vi = int(vf)
		if float64(vi) != vf {
			return 0, missingOrInvalid()
		}
	}

	return vi, nil
}

func String(data interface{}, name string) string {
	s, err := GetString(data, name)
	if err != nil {
		panic(err)
	}

	return s
}

func GetString(data interface{}, name string) (string, error) {
	switch dt := data.(type) {
	case map[string]interface{}:
		v, ok := dt[name].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid field: %s", name)
		}

		return v, nil
	case Struct:
		v, ok := dt[name].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid field: %s", name)
		}

		return v, nil
	default:
		return "", fmt.Errorf("not a struct")
	}
}
