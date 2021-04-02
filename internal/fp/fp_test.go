package fp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_readFloat(t *testing.T) {
	mantissa, exp, neg, trunc, i, ok := readFloat([]byte(`12.345e-9999999`))
	fmt.Println("mantissa", mantissa)
	fmt.Println("exp", exp)
	fmt.Println("neg", neg)
	fmt.Println("trunc", trunc)
	fmt.Println("i", i)
	fmt.Println("ok", ok)
}

func jsonEquiv(data []byte) (val float64, p int, err error) {
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

func testAgainstJSON(t *testing.T, input string) {
	t.Helper()

	want, wantOffset, err := jsonEquiv([]byte(input))
	wantErr := err != nil

	got, gotOffset, gotErr := ParseJSONFloatPrefix([]byte(input))
	if wantErr {
		assert.Error(t, gotErr)
	} else {
		assert.NoError(t, gotErr)
	}
	assert.Equal(t, wantOffset, gotOffset)
	assert.Equal(t, want, got)
}

func Benchmark_readFloat(b *testing.B) {
	data := []byte(`11111.001111111789`)
	var ok bool
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, _, _, _, _, ok = readFloat(data)
	}
	assert.True(b, ok)
}

func BenchmarkParseJSONFloatPrefix(b *testing.B) {
	data := []byte(`11111.001111111789`)
	b.SetBytes(int64(len(data)))
	var err error
	for i := 0; i < b.N; i++ {
		_, _, err = ParseJSONFloatPrefix(data)
	}
	assert.NoError(b, err)
}

func TestParseJSONFloatPrefix(t *testing.T) {
	for _, s := range []string{
		`"foo"`,
		`"foo`,
		" 1",
		"1e4",
		"1e15",
		"1e23",
		"1000000000000001e3",
		"0132311002552003023566733115",
		"-0132311002552003023566733115",
		"1.",
		".",
		"1.1.1",
		"1.e2",
		"0.0001",
		".0001",
		"00",
		"",
		"0x1",
		"0",
		"1",
		"+1",
		"0.0001",
		"1e-100",
		"100000000000000000000000",
		"100000000000000000000001",
		"100000000000000008388608",
		"100000000000000016777215",
		"100000000000000016777216",
		"-1",
		"-0.1",
		"-0",
		"1E-20",
		"625e-3",
		"0e+01234567890123456789",
		"0e291",
		"0e292",
		"0e347",
		"0e348",
		"-0e291",
		"-0e292",
		"-0e347",
		"-0e348",
		"1.7976931348623157e308",
		"-1.7976931348623157e308",
		"1.7976931348623159e308",
		"-1.7976931348623159e308",
		"1.7976931348623158e308",
		"-1.7976931348623158e308",
		"1.797693134862315808e308",
		"-1.797693134862315808e308",
		"1e308",
		"2e308",
		"1e309",
		"1e310",
		"-1e310",
		"1e400",
		"-1e400",
		"1e400000",
		"-1e400000",
		"1e-305",
		"1e-306",
		"1e-307",
		"1e-308",
		"1e-309",
		"1e-310",
		"1e-322",
		"5e-324",
		"4e-324",
		"3e-324",
		"2e-324",
		"1e-350",
		"1e-400000",
	} {
		t.Run(s, func(t *testing.T) {
			testAgainstJSON(t, s)
		})
	}
}
