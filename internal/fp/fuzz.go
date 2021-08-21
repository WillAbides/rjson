// +build gofuzz

package fp

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func fuzzOracle(data []byte) (val float64, p int, err error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty")
	}
	switch data[0] {
	case '\t', '\r', '\n', ' ':
		return 0, 0, fmt.Errorf("whitespace prefix")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	tkn, err := decoder.Token()
	if err != nil {
		return 0, int(decoder.InputOffset()), err
	}
	var ok bool
	val, ok = tkn.(float64)
	if !ok {
		return 0, 0, fmt.Errorf("not a number")
	}
	return val, int(decoder.InputOffset()), nil
}

// RunFuzz runs a fuzz test but returns an error instead of panicking
func RunFuzz(data []byte) (int, error) {
	want, wantOffset, err := fuzzOracle(data)
	wantErr := err != nil
	got, gotOffset, gotErr := ParseJSONFloatPrefix(data)
	if wantErr {
		if gotErr == nil {
			return 0, fmt.Errorf("missed an error")
		}
	} else {
		if gotErr != nil {
			return 0, fmt.Errorf("unexpected error")
		}
	}
	if want != got {
		return 0, fmt.Errorf("wanted %v but got %v", want, got)
	}
	if wantOffset != gotOffset {
		return 0, fmt.Errorf("wanted offset %v but got %v", wantOffset, gotOffset)
	}
	return 0, nil
}

// Fuzz is for running go-fuzz tests
func Fuzz(data []byte) int {
	n, err := RunFuzz(data)
	if err != nil {
		panic(err.Error())
	}
	return n
}
