package rjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:generate go run ./internal/testgen

func TestReadNull(t *testing.T) {
	t.Parallel()
	testReadChecker(t, readNullChecker, true)
}

func TestReadUint64(t *testing.T) {
	data := []byte(`83640
}`)
	val, p, err := ReadUint64(data)
	assert.NoError(t, err)
	assert.EqualValues(t, 5, p)
	assert.EqualValues(t, 83640, val)
}

func readNullChecker(t *testing.T, data []byte) int {
	t.Helper()
	p, err := ReadNull(data)
	var umVal interface{}
	umOffset, umNull, umErr := getUnmarshalResult(data, &umVal)
	if umErr != nil || !umNull {
		assert.Error(t, err)
		return 0
	}
	assert.NoError(t, err)
	assert.Equal(t, umOffset, p)
	return p
}

func testReadChecker(t *testing.T, checker func(t *testing.T, data []byte) int, recurse bool) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(t, file)
		if recurse {
			runChecker(t, data, checker)
		} else {
			checker(t, data)
		}
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func runChecker(t *testing.T, data []byte, checker func(t *testing.T, data []byte) int) {
	t.Helper()
	handler := &simpleValueHandler{
		simpleValueHandler: ArrayValueHandlerFunc(func(data []byte) (int, error) {
			t.Helper()
			p := checker(t, data)
			return p, nil
		}),
	}
	_, err := handler.HandleArrayValue(data)
	assert.NoError(t, err)
}
