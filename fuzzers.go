package rjson

import (
	"encoding/json"
	"fmt"
)

type fuzzer struct {
	name string
	fn   func([]byte) (int, error)
}

var fuzzers = []fuzzer{
	{name: "fuzzReadFloat64", fn: fuzzReadFloat64},
	{name: "fuzzReadUint64", fn: fuzzReadUint64},
	{name: "fuzzReadUint32", fn: fuzzReadUint32},
	{name: "fuzzReadUint", fn: fuzzReadUint},
	{name: "fuzzReadInt64", fn: fuzzReadInt64},
	{name: "fuzzReadInt32", fn: fuzzReadInt32},
	{name: "fuzzReadInt", fn: fuzzReadInt},
	{name: "fuzzReadString", fn: fuzzReadString},
	{name: "fuzzReadStringBytes", fn: fuzzReadStringBytes},
	{name: "fuzzReadBool", fn: fuzzReadBool},
	{name: "fuzzReadNull", fn: fuzzReadNull},
	{name: "fuzzSkipValue", fn: fuzzSkipValue},
	{name: "fuzzValid", fn: fuzzValid},
	{name: "fuzzNextToken", fn: fuzzNextToken},
	{name: "fuzzReadArray", fn: fuzzReadArray},
	{name: "fuzzReadObject", fn: fuzzReadObject},
	{name: "fuzzReadValue", fn: fuzzReadValue},

	{name: "fuzzHandleArrayValues", fn: fuzzHandleArrayValues},
	{name: "fuzzHandleObjectValues", fn: fuzzHandleObjectValues},
}

func fuzzHandleArrayValues(data []byte) (int, error) {
	var buf Buffer

	validArray := false
	valid := Valid(data, &buf)
	if valid {
		tkn, _, err := NextTokenType(data)
		validArray = err == nil && tkn == ArrayStartType
	}

	for name, handlerFunc := range arrayHandlers {
		wantNoErr := validArray && (name == "alwaysZero" || name == "skipValue")
		_, err := HandleArrayValues(data, handlerFunc, &buf)
		if wantNoErr && err != nil {
			return 0, err
		}

		for i := 2; i <= 4; i++ {
			h := nCallArrayValueHandler{
				handler: handlerFunc,
				n:       i,
			}
			_, err = HandleArrayValues(data, &h, nil)
			if wantNoErr && err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}

func fuzzHandleObjectValues(data []byte) (int, error) {
	var buf Buffer
	for _, handlerFunc := range arrayHandlers {
		hh := ObjectValueHandlerFunc(func(_, data []byte) (p int, err error) {
			return handlerFunc(data)
		})
		_, err := HandleObjectValues(data, hh, &buf)
		_ = err //nolint:errcheck // just checking for panic

		for i := 2; i <= 4; i++ {

			h := nCallArrayValueHandler{
				handler: handlerFunc,
				n:       i,
			}
			hh = func(_, data []byte) (p int, err error) {
				return h.HandleArrayValue(data)
			}
			_, err := HandleObjectValues(data, hh, nil)
			_ = err //nolint:errcheck // just checking for panic
		}
	}
	return 0, nil
}

func fuzzReadUint64(data []byte) (int, error) {
	want, wantP, wantErr := readUint64Compat(data)
	got, gotP, gotErr := ReadUint64(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadUint32(data []byte) (int, error) {
	want, wantP, wantErr := readUint32Compat(data)
	got, gotP, gotErr := ReadUint32(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadInt64(data []byte) (int, error) {
	want, wantP, wantErr := readInt64Compat(data)
	got, gotP, gotErr := ReadInt64(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadInt32(data []byte) (int, error) {
	want, wantP, wantErr := readInt32Compat(data)
	got, gotP, gotErr := ReadInt32(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadUint(data []byte) (int, error) {
	want, wantP, wantErr := readUintCompat(data)
	got, gotP, gotErr := ReadUint(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadInt(data []byte) (int, error) {
	want, wantP, wantErr := readIntCompat(data)
	got, gotP, gotErr := ReadInt(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadFloat64(data []byte) (int, error) {
	want, wantP, wantErr := readFloat64Compat(data)
	got, gotP, gotErr := ReadFloat64(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadString(data []byte) (int, error) {
	want, wantP, wantErr := readStringCompat(data)
	got, gotP, gotErr := ReadString(data, nil)
	got = StdLibCompatibleString(got)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	return 0, err
}

func fuzzReadStringBytes(data []byte) (int, error) {
	want, wantP, wantErr := readStringBytesCompat(data)
	gotBytes, gotP, gotErr := ReadStringBytes(data, nil)
	got := StdLibCompatibleString(string(gotBytes))
	err := checkFuzzResults(string(want), got, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	return 0, err
}

func fuzzReadBool(data []byte) (int, error) {
	want, wantP, wantErr := readBoolCompat(data)
	got, gotP, gotErr := ReadBool(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadNull(data []byte) (int, error) {
	wantP, wantErr := readNullCompat(data)
	gotP, gotErr := ReadNull(data)
	err := checkFuzzResults(nil, nil, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzSkipValue(data []byte) (int, error) {
	wantP, wantErr := skipValueCompat(data)
	var buf Buffer
	gotP, gotErr := SkipValue(data, &buf)
	err := checkFuzzResults(nil, nil, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	// try again with nil buffer
	gotP, gotErr = SkipValue(data, &buf)
	err = checkFuzzResults(nil, nil, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzValid(data []byte) (int, error) {
	want := json.Valid(data)
	var buf Buffer
	got := Valid(data, &buf)
	err := checkFuzzResults(want, got, 0, 0, nil, nil)
	if err != nil {
		return 0, err
	}
	// try again with nil buffer
	got = Valid(data, nil)
	err = checkFuzzResults(want, got, 0, 0, nil, nil)
	return 0, err
}

func fuzzNextToken(data []byte) (int, error) {
	want, wantP, wantErr := nextTokenCompat(data)
	got, gotP, gotErr := NextToken(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzNextTokenType(data []byte) (int, error) {
	want, wantP, wantErr := nextTokenTypeCompat(data)
	got, gotP, gotErr := NextTokenType(data)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadArray(data []byte) (int, error) {
	want, wantP, wantErr := readArrayCompat(data)
	got, gotP, gotErr := ReadArray(data)
	got = StdLibCompatibleSlice(got)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	// try again with a ValueReader
	got, gotP, gotErr = (&ValueReader{}).ReadArray(data)
	got = StdLibCompatibleSlice(got)
	err = checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadObject(data []byte) (int, error) {
	want, wantP, wantErr := readObjectCompat(data)
	got, gotP, gotErr := ReadObject(data)
	got = StdLibCompatibleMap(got)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	// try again with a ValueReader
	got, gotP, gotErr = (&ValueReader{}).ReadObject(data)
	got = StdLibCompatibleMap(got)
	err = checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func fuzzReadValue(data []byte) (int, error) {
	want, wantP, wantErr := readValueCompat(data)
	got, gotP, gotErr := ReadValue(data)
	got = stdLibCompatibleValue(got)
	err := checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	if err != nil {
		return 0, err
	}
	// try again with a ValueReader
	got, gotP, gotErr = (&ValueReader{}).ReadValue(data)
	got = stdLibCompatibleValue(got)
	err = checkFuzzResults(want, got, wantP, gotP, wantErr, gotErr)
	return 0, err
}

func checkFuzzResults(want, got interface{}, wantP, gotP int, wantErr, gotErr error) error {
	err := checkFuzzErrors(wantErr, gotErr)
	if err != nil || wantErr != nil {
		return err
	}
	err = fuzzCompare(want, got)
	if err != nil {
		return err
	}
	if gotP != wantP {
		return fmt.Errorf("expected p=%d, but got p=%d", wantP, gotP)
	}
	return nil
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
