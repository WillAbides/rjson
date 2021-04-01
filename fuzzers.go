package rjson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

//go:generate go run ./internal/fuzzgen

type fuzzer struct {
	name string
	fn   func([]byte) (int, error)
}

var fuzzers = append(generatedFuzzers,
	fuzzer{name: "fuzzSkip", fn: fuzzSkip},
	fuzzer{name: "fuzzNextToken", fn: fuzzNextToken},
	fuzzer{name: "fuzzValid", fn: fuzzValid},
	fuzzer{name: "fuzzHandleArrayValues", fn: fuzzHandleArrayValues},
	fuzzer{name: "fuzzHandleObjectValues", fn: fuzzHandleObjectValues},
)

func fuzzValid(data []byte) (int, error) {
	want := json.Valid(data)
	got := Valid(data, nil)
	if want && !got {
		return 0, fmt.Errorf("expected valid but got invalid")
	}
	if got && !want {
		return 0, fmt.Errorf("expected invalid but got valid")
	}
	return 0, nil
}

func fuzzSkip(data []byte) (int, error) {
	var buf Buffer
	skippedBytes, err := SkipValue(data, &buf)
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

func fuzzHandleArrayValues(data []byte) (int, error) {
	var buf Buffer
	for _, handlerFunc := range arrayHandlers {
		_, err := HandleArrayValues(data, handlerFunc, &buf)
		_ = err //nolint:errcheck // just checking for panic

		for i := 2; i <= 4; i++ {

			h := nCallArrayValueHandler{
				handler: handlerFunc,
				n:       i,
			}
			_, err := HandleArrayValues(data, &h, nil)
			_ = err //nolint:errcheck // just checking for panic
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
			hh = ObjectValueHandlerFunc(func(_, data []byte) (p int, err error) {
				return h.HandleArrayValue(data)
			})
			_, err := HandleObjectValues(data, hh, nil)
			_ = err //nolint:errcheck // just checking for panic
		}
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
