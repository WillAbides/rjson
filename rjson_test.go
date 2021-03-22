package rjson

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These are the files in testdata.  They all come directly from github.com/pkg/json, but I'm keeping the comments
// about where they are originally from.
var jsonTestFiles = []string{

	// from https://github.com/miloyip/nativejson-benchmark
	"canada.json", "citm_catalog.json", "twitter.json", "code.json",

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	"example.json",

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	"sample.json",
}

func Test_fuzzers(t *testing.T) {
	corpusDir := filepath.FromSlash(`testdata/fuzz/corpus`)
	if !fileExists(t, filepath.FromSlash(corpusDir)) {
		t.Skip()
	}

	for _, td := range []struct {
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
		t.Run(td.name, func(t *testing.T) {
			dir, err := ioutil.ReadDir(corpusDir)
			require.NoError(t, err)
			for _, info := range dir {
				data, err := ioutil.ReadFile(filepath.Join(corpusDir, info.Name()))
				require.NoError(t, err)
				_, err = td.fuzzer(data)
				assert.NoErrorf(t, err, "error from file: %s\n with data:\n%s", filepath.Join(corpusDir, info.Name()), string(data))
			}
		})
	}
}

func TestUnmarshalFace(t *testing.T) {
	for _, s := range jsonTestFiles {
		t.Run(s, func(t *testing.T) {
			data := getTestdataJSONGz(t, s)
			assertIfaceUnmarshalsSame(t, data)
		})
	}

	// various values that fuzz has crashed on in the past
	for _, fuzzy := range []string{
		`"浱up蔽Cr"`,
		"{\"\":\"0\xc9\"}",
		"\"\x90\"",
		"{\"\x14\":0}",
		"0\xff",
		"{\"a\": \"b\", \"\xbf\":\"\"}",
		"\"\x1b\"",
		"{\"a\":\"\x1b\"}",
		"{\"\x1b\":true}",
		"\"\x1f\"",
		"\"\\'\"",
		"80000000000000000000",
		"999999999999999999",
	} {
		t.Run(fuzzy, func(t *testing.T) {
			assertIfaceUnmarshalsSame(t, []byte(fuzzy))
		})
	}
}

// asserts that ifaceUnmarshaller{} gets the same result as (*json.Decoder).Decode
// first it must fix an issue with *json.Decoder decoding some chars to RuneError
func assertIfaceUnmarshalsSame(t *testing.T, data []byte) bool {
	handler := &ifaceUnmarshaller{}
	_, gotErr := handler.HandleAnyValue(data)
	var want interface{}
	wantErr := json.NewDecoder(bytes.NewReader(data)).Decode(&want)
	if wantErr != nil {
		return assert.Errorf(t, gotErr, "we got no error but encoding/json got: %v", wantErr)
	}
	if !assert.NoError(t, gotErr) {
		return false
	}
	got, want := removeJSONRuneError(handler.val, want)
	return assert.Equal(t, want, got)
}

func TestBuffer_SkipValue(t *testing.T) {
	buf := new(Buffer)
	for _, s := range jsonTestFiles {
		t.Run(s, func(t *testing.T) {
			data := getTestdataJSONGz(t, s)
			skippedBytes, err := buf.SkipValue(data)
			gotValid := err == nil
			skippedData := bytes.TrimLeft(data[:skippedBytes], " \t\n\r")
			wantValid := json.Valid(skippedData)
			require.Equal(t, wantValid, gotValid)
		})
	}

	t.Run("invalid json", func(t *testing.T) {
		for _, s := range invalidJSON {
			data := []byte(s)
			skippedBytes, err := buf.SkipValue(data)
			gotValid := err == nil
			skippedData := bytes.TrimLeft(data[:skippedBytes], " \t\n\r")
			wantValid := json.Valid(skippedData)
			require.Equalf(t, wantValid, gotValid, "not equal for %q", s)
		}
	})
}

func TestSkipValue(t *testing.T) {
	for _, s := range jsonTestFiles {
		t.Run(s, func(t *testing.T) {
			data := getTestdataJSONGz(t, s)
			skippedBytes, err := SkipValue(data)
			gotValid := err == nil
			skippedData := bytes.TrimLeft(data[:skippedBytes], " \t\n\r")
			wantValid := json.Valid(skippedData)
			require.Equal(t, wantValid, gotValid)
		})
	}

	t.Run("invalid json", func(t *testing.T) {
		for _, s := range invalidJSON {
			data := []byte(s)
			skippedBytes, err := SkipValue(data)
			gotValid := err == nil
			skippedData := bytes.TrimLeft(data[:skippedBytes], " \t\n\r")
			wantValid := json.Valid(skippedData)
			assert.Equalf(t, wantValid, gotValid, "not equal for %q", s)
		}
	})
}

func gunzipTestJSON(t testing.TB, filename string) string {
	t.Helper()
	targetDir := filepath.Join("testdata", "tmp")
	err := os.MkdirAll(targetDir, 0o700)
	require.NoError(t, err)
	target := filepath.Join(targetDir, filename)
	if fileExists(t, target) {
		return target
	}
	src := filepath.Join("testdata", filename+".gz")
	f, err := os.Open(src)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()
	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	buf, err := ioutil.ReadAll(gz)
	require.NoError(t, err)
	err = ioutil.WriteFile(target, buf, 0o600)
	require.NoError(t, err)
	return target
}

func getTestdataJSONGz(t testing.TB, path string) []byte {
	t.Helper()
	filename := gunzipTestJSON(t, path)
	got, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return got
}

func fileExists(t testing.TB, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	require.NoError(t, err)
	return true
}

func runChecker(t *testing.T, data []byte, checker func(t *testing.T, data []byte) int) {
	t.Helper()
	handler := &simpleValueHandler{
		simpleValueHandler: ValueHandlerFunc(func(data []byte) (int, error) {
			t.Helper()
			p := checker(t, data)
			return p, nil
		}),
	}
	_, err := handler.HandleValue(data)
	require.NoError(t, err)
}

func getUnmarshalResult(data []byte, v interface{}) (offset int, null bool, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()
	offset = int(decoder.InputOffset())
	if err != nil {
		return offset, false, err
	}
	if token == nil {
		return offset, true, nil
	}
	err = json.Unmarshal(data[:offset], v)
	return offset, false, err
}

func assertMatchesUnmarshal(t *testing.T, data []byte, got interface{}, gotP int, gotErr error) int {
	umVal := reflect.New(reflect.TypeOf(got)).Interface()
	gotAnError := gotErr != nil
	umOffset, umNull, umErr := getUnmarshalResult(data, &umVal)
	wantErr := umNull || umErr != nil

	pp := gotP
	if umOffset > pp {
		pp = umOffset
	}
	msgData := "raw json: " + string(data[:pp])

	assert.Equal(t, wantErr, gotAnError, msgData)
	wantVal := reflect.Indirect(reflect.ValueOf(umVal)).Interface()
	assert.EqualValues(t, wantVal, got, msgData)
	if !wantErr {
		assert.Equal(t, umOffset, gotP, msgData)
	}
	if gotAnError {
		return 0
	}
	return gotP
}

func TestReadStringBytes(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadStringBytes(data, nil)
		return assertMatchesUnmarshal(t, data, string(got), p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
}

func TestReadString(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadString(data, nil)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
}

func TestReadUint64(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadUint64(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadUint(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadUint(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadUint32(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadUint32(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadInt(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadInt(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadInt32(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadInt32(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadInt64(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadInt64(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadFloat64(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadFloat64(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadBool(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
		t.Helper()
		got, p, err := ReadBool(data)
		return assertMatchesUnmarshal(t, data, got, p, err)
	}

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestReadNull(t *testing.T) {
	t.Parallel()

	checker := func(t *testing.T, data []byte) int {
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

	for _, file := range jsonTestFiles {
		runChecker(t, getTestdataJSONGz(t, file), checker)
	}
	for _, s := range invalidJSON {
		checker(t, []byte(s))
	}
}

func TestNextToken(t *testing.T) {
	t.Parallel()
	for _, td := range []struct {
		data string
		tkn  byte
		p    int
		err  string
	}{
		{data: `"foo"`, tkn: '"', p: 1},
		{data: `"`, tkn: '"', p: 1},
		{data: "\n\ntrue", tkn: 't', p: 3},
		{data: ``, err: "EOF"},
		{data: ` `, p: 1, err: "EOF"},
		{data: ` asdf `, tkn: 'a', p: 2, err: "no valid json token found"},
		{data: `asdf `, tkn: 'a', p: 1, err: "no valid json token found"},
	} {
		t.Run(td.data, func(t *testing.T) {
			tkn, p, err := NextToken([]byte(td.data))
			if td.err != "" {
				require.EqualError(t, err, td.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, string(td.tkn), string(tkn))
			require.Equal(t, td.p, p)
		})
	}
}

func TestNextTokenType(t *testing.T) {
	for _, td := range []struct {
		data string
		tkn  TokenType
		p    int
		err  string
	}{
		{data: `"`, tkn: StringType, p: 1},
		{data: "true", tkn: TrueType, p: 1},
		{data: "false", tkn: FalseType, p: 1},
		{data: "null", tkn: NullType, p: 1},
		{data: "{", tkn: ObjectStartType, p: 1},
		{data: "}", tkn: ObjectEndType, p: 1},
		{data: "[", tkn: ArrayStartType, p: 1},
		{data: "]", tkn: ArrayEndType, p: 1},
		{data: ",", tkn: CommaType, p: 1},
		{data: ":", tkn: ColonType, p: 1},
		{data: "0", tkn: NumberType, p: 1},
		{data: ``, err: "EOF"},
		{data: ` `, p: 1, err: "EOF"},
		{data: ` asdf `, p: 2},
		{data: `asdf `, p: 1},
	} {
		t.Run(td.data, func(t *testing.T) {
			tkn, p, err := NextTokenType([]byte(td.data))
			if td.err != "" {
				require.EqualError(t, err, td.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, string(td.tkn), string(tkn))
			require.Equal(t, td.p, p)
		})
	}

	fileCounts := map[string]int{
		"canada.json":       334373,
		"citm_catalog.json": 135990,
		"twitter.json":      55263,
		"code.json":         396293,
		"example.json":      1297,
	}
	for filename, wantCount := range fileCounts {
		t.Run(filename, func(t *testing.T) {
			assert.Equal(t, wantCount, countTokens(getTestdataJSONGz(t, filename)))
		})
	}
}

func countTokens(data []byte) int {
	var count int
	buf := new(Buffer)
	for {
		tp, p, err := NextTokenType(data)
		if err != nil {
			break
		}
		count++
		data = data[p-1:]
		switch tp {
		case NullType, StringType, TrueType, FalseType, NumberType:
			p, err = buf.SkipValue(data)
			if err != nil {
				return count
			}
			data = data[p:]
		case InvalidType:
			fmt.Println(data[0])
			count--
			if len(data) > 0 {
				data = data[1:]
			}
		default:
			if len(data) > 0 {
				data = data[1:]
			}
		}
	}
	return count
}

type simpleValueHandler struct {
	buffer             Buffer
	simpleValueHandler ValueHandler
}

func (h *simpleValueHandler) HandleObjectValue(fieldname, data []byte) (int, error) {
	return h.HandleValue(data)
}

func (h *simpleValueHandler) HandleValue(data []byte) (int, error) {
	tknType, p, err := NextTokenType(data)
	if err != nil {
		return p, err
	}
	p--
	data = data[p:]
	var pp int
	switch tknType {
	case ObjectStartType:
		pp, err = h.buffer.HandleObjectValues(data, h)
	case ArrayStartType:
		pp, err = h.buffer.HandleArrayValues(data, h)
	default:
		pp, err = h.simpleValueHandler.HandleValue(data)
	}
	return pp, err
}

// examples taken from github.com/pkg/json
var invalidJSON = []string{
	`[`,
	`{"":2`,
	`[[[[]]]`,
	`{"`,
	`{"":` + "\n" + `}`,
	`{{"key": 1}: 2}}`,
	`{1: 1}`,
	`[[],[], [[]],�[[]]]`,
	`+`,
	`,`,
	`1.e1`,
	`{"a":"b":"c"}`,
	`{"test"::"input"}`,
	`e1`,
	`-.1e-1`,
	`123.`,
	`--123`,
	`.1`,
	`0.1e`,
	"{\"foo\": \"\x14\"}",
	"{\"\x14\":0}",
}
