// +build gofuzz

package rjson

import (
	"fmt"

	"github.com/willabides/rjson/internal/fp"
)

// Fuzz is for running go-fuzz tests
func Fuzz(data []byte) int {
	score := 0
	allFuzzers := append(fuzzers, fuzzer{
		name: "fp.RunFuzz",
		fn:   fp.RunFuzz,
	})
	for _, fd := range allFuzzers {
		d := make([]byte, len(data))
		copy(d, data)
		s, err := fd.fn(d)
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
