// +build gofuzz

package rjson

import (
	"fmt"
)

// Fuzz is for running go-fuzz tests
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
		{name: "fuzzValid", fuzzer: fuzzValid},
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
