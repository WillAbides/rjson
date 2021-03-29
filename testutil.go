package rjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type fuzzer struct {
	name string
	fn   func([]byte) (int, error)
}

var fuzzers = append(generatedFuzzers,
	fuzzer{name: "fuzzSkip", fn: fuzzSkip},
	fuzzer{name: "fuzzNextToken", fn: fuzzNextToken},
	fuzzer{name: "fuzzValid", fn: fuzzValid},
)

// removeJSONRuneError because encoding/json incorrectly unmarshals some characters to RuneError
// when we see that rjson differs from encoding/json change both values to ""  before comparing
func removeJSONRuneError(rjsonVal, jsonVal interface{}) (rvRes, jvRes interface{}) {
	var rvMap, jvMap map[string]interface{}
	var ok bool
	rvMap, ok = rjsonVal.(map[string]interface{})
	if ok {
		jvMap, ok = jsonVal.(map[string]interface{})
		if !ok {
			return rjsonVal, jsonVal
		}
		var jvv interface{}
		for k, rvv := range rvMap {
			if strings.ContainsRune(k, utf8.RuneError) {
				delete(rvMap, k)
				continue
			}
			jvv, ok = jvMap[k]
			if !ok {
				delete(rvMap, k)
				continue
			}
			rvMap[k], jvMap[k] = removeJSONRuneError(rvv, jvv)
		}
		for k := range jvMap {
			if !strings.ContainsRune(k, utf8.RuneError) {
				continue
			}
			_, ok = rvMap[k]
			if !ok {
				delete(jvMap, k)
			}
		}
		return rvMap, jvMap
	}

	var rvSlice, jvSlice []interface{}
	rvSlice, ok = rjsonVal.([]interface{})
	if ok {
		jvSlice, ok = jsonVal.([]interface{})
		if !ok {
			return rjsonVal, jsonVal
		}
		for i := range rvSlice {
			if i >= len(jvSlice) {
				break
			}
			rvSlice[i], jvSlice[i] = removeJSONRuneError(rvSlice[i], jvSlice[i])
		}
		return rvSlice, jvSlice
	}

	var rvStr, jvStr string
	rvStr, ok = rjsonVal.(string)
	if ok {
		jvStr, ok = jsonVal.(string)
		if !ok {
			return rjsonVal, jsonVal
		}
		if strings.ContainsRune(jvStr, utf8.RuneError) {
			jvStr = ""
			rvStr = ""
		}
		return jvStr, rvStr
	}

	return rjsonVal, jsonVal
}

func fuzzValid(data []byte) (int, error) {
	want := json.Valid(data)
	got := Valid(data)
	if want && !got {
		return 0, fmt.Errorf("expected valid but got invalid")
	}
	if got && !want {
		return 0, fmt.Errorf("expected invalid but got valid")
	}
	return 0, nil
}

func fuzzSkip(data []byte) (int, error) {
	skippedBytes, err := SkipValue(data)
	gotValid := err == nil
	skippedData := data
	if skippedBytes < len(skippedData) {
		skippedData = skippedData[:skippedBytes]
	}
	wantValid := json.Valid(skippedData)
	if wantValid && !gotValid {
		return 0, fmt.Errorf("failed to skip valid json. error: %v", err)
	}
	if !wantValid && gotValid {
		return 0, fmt.Errorf("failed to detect invalid json")
	}
	return 0, nil
}

func checkFuzzErrors(wantErr, gotErr error) error {
	if wantErr != nil {
		if gotErr == nil {
			return fmt.Errorf("we got no error but json got: %v", wantErr)
		}
		return nil
	}
	if gotErr != nil {
		return fmt.Errorf("json got no error but we did: %v", gotErr)
	}
	return nil
}

func fuzzNextToken(data []byte) (int, error) {
	got, _, gotErr := NextToken(data)

	want, wantErr := json.NewDecoder(bytes.NewReader(data)).Token()
	if wantErr != nil {
		return 0, nil
	}
	if gotErr != nil {
		return 0, fmt.Errorf("json got no error but we did: %v", gotErr)
	}
	var wantType TokenType
	switch w := want.(type) {
	case json.Delim:
		wantType = tokenTypes[w]
	case bool:
		wantType = TrueType
		if !w {
			wantType = FalseType
		}
	case float64:
		wantType = NumberType
	case string:
		wantType = StringType
	case nil:
		wantType = NullType
	}
	gotType := tokenTypes[got]
	if wantType != gotType {
		return 0, fmt.Errorf("got wrong token type. wanted %s but got %s", wantType, gotType)
	}
	return 0, nil
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
