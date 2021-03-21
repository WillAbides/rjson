package rjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func Fuzz(data []byte) int {
	score := 0

	for _, fd := range []struct {
		name   string
		fuzzer fuzzer
	}{
		{name: "fuzzIfaceUnmarshaller", fuzzer: fuzzIfaceUnmarshaller},
		{name: "fuzzSkip", fuzzer: fuzzSkip},
		{name: "fuzzNextToken", fuzzer: fuzzNextToken},
		{name: "fuzzReadUint64", fuzzer: fuzzReadUint64},
		{name: "fuzzReadInt64", fuzzer: fuzzReadInt64},
		{name: "fuzzReadUint32", fuzzer: fuzzReadUint32},
		{name: "fuzzReadInt32", fuzzer: fuzzReadInt32},
		{name: "fuzzReadInt", fuzzer: fuzzReadInt},
		{name: "fuzzReadUint", fuzzer: fuzzReadUint},
		{name: "fuzzReadFloat64", fuzzer: fuzzReadFloat64},
	} {
		d := make([]byte, len(data))
		copy(d, data)
		s, err := fd.fuzzer(d)
		if err != nil {
			panic(fmt.Sprintf("%s\n%v", fd.name, err))
		}
		switch s {
		case -1:
			return -1
		case 1:
			score = 1
		}
	}
	return score
}

type fuzzer func([]byte) (int, error)

func fuzzIfaceUnmarshaller(data []byte) (int, error) {
	var jsonVal interface{}
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&jsonVal)

	handler := &ifaceUnmarshaller{}
	_, gotErr := handler.HandleAnyValue(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	if wantErr != nil {
		return 0, nil
	}
	gotVal, wantVal := removeJSONRuneError(handler.val, jsonVal)
	err = ifaceCompare(wantVal, gotVal, []string{"ROOT"})
	if err != nil {
		return 0, err
	}
	return 1, nil
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

var (
	startsWithNull = regexp.MustCompile(`^[\t\r\n ]*null`)
	startsWith00   = regexp.MustCompile(`^[\t\r\n ]*-?0[0-9eE]`)
)

func checkValidNumberFuzzers(data []byte) bool {
	if startsWithNull.Match(data) {
		return false
	}
	if startsWith00.Match(data) {
		return false
	}
	return true
}

func fuzzPrepReadData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	return data[countWhitespace(data):]
}

func fuzzReadUint64(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want uint64
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadUint64(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadInt64(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want int64
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadInt64(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadUint32(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want uint32
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadUint32(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadInt32(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want int32
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadInt32(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadFloat64(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want float64
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadFloat64(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadUint(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want uint
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadUint(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
	}
	return 0, nil
}

func fuzzReadInt(data []byte) (int, error) {
	data = fuzzPrepReadData(data)
	if !checkValidNumberFuzzers(data) {
		return 0, nil
	}
	var want int
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	got, _, gotErr := ReadInt(data)
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return 0, err
	}
	if got != want {
		return 0, fmt.Errorf("expected %v but got %v", want, got)
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
