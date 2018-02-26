package netreduce

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/netreduce/netreduce/data"
)

// TODO: this needs loop protection if used for real

func Wrap(d interface{}) interface{} {
	switch dt := d.(type) {
	case map[string]interface{}:
		s := data.Struct(dt)
		for key, value := range s {
			s[key] = Wrap(value)
		}

		return s
	case []interface{}:
		l := data.List(dt)
		for i, value := range l {
			l[i] = Wrap(value)
		}

		return l
	case float64:
		intd := int(dt)
		if dt == float64(intd) {
			return intd
		}

		return d
	default:
		return d
	}
}

func GetJSON(u string) (interface{}, error) {
	rsp, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", rsp.Status)
	}

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(b, &data)
	return data, err
}

func Equal(left, right interface{}) bool {
	wl, wr := Wrap(left), Wrap(right)

	switch dtl := wl.(type) {
	case data.Struct:
		switch dtr := wr.(type) {
		case data.Struct:
			if len(dtl) != len(dtr) {
				return false
			}

			for key, vl := range dtl {
				vr, ok := dtr[key]
				if !ok {
					return false
				}

				if !Equal(vl, vr) {
					return false
				}
			}

			return true
		default:
			return false
		}
	case data.List:
		switch dtr := wr.(type) {
		case data.List:
			if len(dtl) != len(dtr) {
				return false
			}

			for index, vl := range dtl {
				vr := dtr[index]
				if !Equal(vl, vr) {
					return false
				}
			}

			return true
		default:
			return false
		}
	default:
		return wl == wr
	}
}
