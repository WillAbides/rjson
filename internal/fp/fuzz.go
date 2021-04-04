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

// Fuzz is for running go-fuzz tests
func Fuzz(data []byte) int {
	want, wantOffset, err := fuzzOracle(data)
	wantErr := err != nil
	got, gotOffset, gotErr := ParseJSONFloatPrefix(data)
	if wantErr {
		if gotErr == nil {
			panic("missed an error")
		}
	} else {
		if gotErr != nil {
			panic("unexpected error")
		}
	}
	if want != got {
		panic(fmt.Sprintf("wanted %v but got %v\n", want, got))
	}
	if wantOffset != gotOffset {
		panic(fmt.Sprintf("wanted offset %v but got %v\n", wantOffset, gotOffset))
	}
	return 0
}
