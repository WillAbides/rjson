package rjson

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var arrayHandlers = map[string]ArrayValueHandlerFunc{
	"alwaysZero":  func([]byte) (p int, err error) { return 0, nil },
	"skipValue":   func(data []byte) (p int, err error) { return SkipValue(data, nil) },
	"alwaysError": func([]byte) (p int, err error) { return 0, fmt.Errorf("error") },
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
	"neg1":          func([]byte) (p int, err error) { return -1, nil },
	"neg100_000":    func([]byte) (p int, err error) { return -100_000, nil },
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

func stdLibCompatibleValue(rjsonVal any) any {
	switch v := rjsonVal.(type) {
	case string:
		return StdLibCompatibleString(v)
	case map[string]any:
		return StdLibCompatibleMap(v)
	case []any:
		return StdLibCompatibleSlice(v)
	default:
		return v
	}
}

type multiPathError []*pathError

func (m multiPathError) Error() string {
	msg := ""
	for _, err := range m {
		msg += err.Error() + "\n"
	}
	return msg
}

type pathError struct {
	path []string
	msg  string
}

func (p *pathError) Error() string {
	return strings.Join(p.path, ".") + ": " + p.msg
}

func newPathErr(path []string, msg string, args ...any) *pathError {
	return &pathError{
		path: path,
		msg:  fmt.Sprintf(msg, args...),
	}
}

func wrongValErr(path []string, a, b any) *pathError {
	return newPathErr(path, "wrong value. wanted %v but got %v", a, b)
}

func wrongTypeErr(path []string, a, b any) *pathError {
	return newPathErr(path, "wrong type. wanted %T but got %T", a, b)
}

func fuzzCompare(want, got any) error {
	switch want.(type) {
	case map[string]any, []any:
		return ifaceCompare(want, got, []string{"ROOT"})
	}
	if got != want {
		return fmt.Errorf("expected %v but got %v", want, got)
	}
	return nil
}

func ifaceCompare(want, got any, path []string) error {
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
	case map[string]any:
		gotVal, ok := got.(map[string]any)
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		var multiErr multiPathError
		for k, wv := range wantVal {
			var gv any
			gv, ok = gotVal[k]
			if !ok {
				multiErr = append(multiErr, newPathErr(append(path, k), "missing map key"))
				continue
			}
			if strings.ContainsRune(k, utf8.RuneError) {
				continue
			}
			err = ifaceCompare(wv, gv, append(path, k))
			var pe *pathError
			if errors.As(err, &pe) {
				multiErr = append(multiErr, pe)
			}
			var multiErr2 multiPathError
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
	case []any:
		gotVal, ok := got.([]any)
		if !ok {
			return wrongTypeErr(path, wantVal, got)
		}
		var multiErr multiPathError
		for i := range wantVal {
			pathElem := fmt.Sprintf(`[%d]`, i)
			if i >= len(gotVal) {
				multiErr = append(multiErr, newPathErr(append(path, pathElem), "missing value"))
			}
			err = ifaceCompare(wantVal[i], gotVal[i], append(path, pathElem))
			var pe *pathError
			if errors.As(err, &pe) {
				multiErr = append(multiErr, pe)
			}
			var multiErr2 multiPathError
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

func dirtyStringBuffer() *[]byte {
	buf := []byte(`this is a dirty buffer`)
	return &buf
}
