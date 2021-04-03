package rjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var arrayHandlers = map[string]ArrayValueHandlerFunc{
	"alwaysZero":  func(data []byte) (p int, err error) { return 0, nil },
	"skipValue":   func(data []byte) (p int, err error) { return SkipValue(data, nil) },
	"alwaysError": func(data []byte) (p int, err error) { return 0, fmt.Errorf("error") },
	"skipHalf": func(data []byte) (p int, err error) {
		value, err := SkipValue(data, nil)
		return value / 2, err
	},
	"skipDouble": func(data []byte) (p int, err error) {
		value, err := SkipValue(data, nil)
		return value * 2, err
	},
	"skipAll":       func(data []byte) (p int, err error) { return len(data), nil },
	"skipAllTimes2": func(data []byte) (p int, err error) { return len(data) * 2, nil },
	"skipAllPlus1":  func(data []byte) (p int, err error) { return len(data) + 1, nil },
	"neg1":          func(data []byte) (p int, err error) { return -1, nil },
	"neg100_000":    func(data []byte) (p int, err error) { return -100_000, nil },
}

type nCallArrayValueHandler struct {
	handler   ArrayValueHandler
	callCount int
	n         int
}

func (h *nCallArrayValueHandler) HandleArrayValue(data []byte) (p int, err error) {
	h.callCount++
	if h.callCount == h.n {
		return h.handler.HandleArrayValue(data)
	}
	return 0, nil
}

func stdLibCompatibleValue(rjsonVal interface{}) interface{} {
	switch v:= rjsonVal.(type) {
	case []byte:
		return StdLibCompatibleString(string(v))
	case string:
		return StdLibCompatibleString(v)
	case map[string]interface{}:
		return StdLibCompatibleMap(v)
	case []interface{}:
		return StdLibCompatibleSlice(v)
	default:
		return v
	}
}

type multiPathErr []*pathErr

func (m multiPathErr) Error() string {
	msg := ""
	for _, err := range m {
		msg += err.Error() + "\n"
	}
	return msg
}

type pathErr struct {
	path []string
	msg  string
}

func (p *pathErr) Error() string {
	return strings.Join(p.path, ".") + ": " + p.msg
}

func newPathErr(path []string, msg string, args ...interface{}) *pathErr {
	return &pathErr{
		path: path,
		msg:  fmt.Sprintf(msg, args...),
	}
}

func wrongValErr(path []string, a, b interface{}) *pathErr {
	return newPathErr(path, "wrong value. wanted %v but got %v", a, b)
}

func wrongTypeErr(path []string, a, b interface{}) *pathErr {
	return newPathErr(path, "wrong type. wanted %T but got %T", a, b)
}

func fuzzCompare(want, got interface{}) error {
	switch want.(type) {
	case map[string]interface{}, []interface{}:
		return ifaceCompare(want, got, []string{"ROOT"})
	}
	if got != want {
		return fmt.Errorf("expected %v but got %v", want, got)
	}
	return nil
}

func ifaceCompare(want, got interface{}, path []string) error {
	var err error
	switch wantVal := want.(type) {
	case string:
		gotVal, ok := got.(string)
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		if wantVal != gotVal {
			return wrongValErr(path, wantVal, gotVal)
		}
		return nil
	case float64:
		bVal, ok := got.(float64)
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		if wantVal != bVal {
			return wrongValErr(path, wantVal, bVal)
		}
		return nil
	case bool:
		bVal, ok := got.(bool)
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		if wantVal != bVal {
			return wrongValErr(path, wantVal, bVal)
		}
		return nil
	case nil:
		if got != nil {
			return newPathErr(path, "wrong value. wanted nil but got %v", got)
		}
		return nil
	case map[string]interface{}:
		gotVal, ok := got.(map[string]interface{})
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		var multiErr multiPathErr
		for k, wv := range wantVal {
			var gv interface{}
			gv, ok = gotVal[k]
			if !ok {
				multiErr = append(multiErr, newPathErr(append(path, k), "missing map key"))
				continue
			}
			err = ifaceCompare(wv, gv, append(path, k))
			var pe *pathErr
			if errors.As(err, &pe) {
				multiErr = append(multiErr, pe)
			}
			var multiErr2 multiPathErr
			if errors.As(err, &multiErr2) {
				multiErr = append(multiErr, multiErr2...)
			}
		}
		for k := range gotVal {
			_, ok = wantVal[k]
			if !ok {
				multiErr = append(multiErr, newPathErr(append(path, k), "extra map key"))
			}
		}
		if len(multiErr) > 0 {
			return multiErr
		}
		return nil
	case []interface{}:
		gotVal, ok := got.([]interface{})
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		var multiErr multiPathErr
		for i := range wantVal {
			pathElem := fmt.Sprintf(`[%d]`, i)
			if i >= len(gotVal) {
				multiErr = append(multiErr, newPathErr(append(path, pathElem), "missing value"))
			}
			err = ifaceCompare(wantVal[i], gotVal[i], append(path, pathElem))
			var pe *pathErr
			if errors.As(err, &pe) {
				multiErr = append(multiErr, pe)
			}
			var multiErr2 multiPathErr
			if errors.As(err, &multiErr2) {
				multiErr = append(multiErr, multiErr2...)
			}
		}
		for i := len(gotVal); i < len(wantVal); i++ {
			pathElem := fmt.Sprintf(`[%d]`, i)
			multiErr = append(multiErr, newPathErr(append(path, pathElem), "extra value"))
		}
		if len(multiErr) > 0 {
			return multiErr
		}
		return nil
	default:
		return newPathErr(path, "unhandled type %T", wantVal)
	}
}

func skipValueEquiv(data []byte) (p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	tkn, err := decoder.Token()
	if err != nil {
		return int(decoder.InputOffset()), err
	}
	delim, ok := tkn.(json.Delim)
	if !ok {
		return int(decoder.InputOffset()), nil
	}
	switch delim {
	case '[', '{':
	default:
		return int(decoder.InputOffset()), fmt.Errorf("invalid json")
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var val interface{}
	err = decoder.Decode(&val)
	return int(decoder.InputOffset()), err
}
