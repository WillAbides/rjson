// +build gofuzz

package rjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func Fuzz(data []byte) int {
	var jsonVal interface{}

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&jsonVal)
	jsonUnmarshalFailed := err != nil

	handler := &ifaceUnmarshaller{}

	_, err = handler.HandleAnyValue(data)

	if err != nil {
		if jsonUnmarshalFailed {
			return 0
		}
		panic("we errored when json.Unmarshal didn't")
	}

	if jsonUnmarshalFailed {
		panic("json.Unmarshal errored and we didn't")
	}
	gotVal, wantVal := removeJSONRuneError(handler.val, jsonVal)
	err = ifaceCompare(wantVal, gotVal, []string{"ROOT"})
	if err != nil {
		panic(err)
	}
	return 1
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
