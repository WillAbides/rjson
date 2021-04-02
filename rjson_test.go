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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These are the files in testdata.
var jsonTestFiles = []string{
	// from querying github's REST api
	"github_user.json",
	"github_repo.json",

	// from json checker test suite
	"jsonchecker_pass1.json",

	// from https://github.com/miloyip/nativejson-benchmark
	"canada.json",
	"citm_catalog.json",
	"twitter.json",
	"code.json",

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	"example.json",

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	"sample.json",
}

func corpusFiles(t *testing.T) []string {
	t.Helper()
	var result []string
	corpusDir := filepath.FromSlash(`testdata/fuzz/corpus`)
	dir, err := ioutil.ReadDir(corpusDir)
	require.NoError(t, err)
	for _, info := range dir {
		result = append(result, filepath.Join(corpusDir, info.Name()))
	}
	return result
}

func Test_fuzzers(t *testing.T) {
	for _, td := range fuzzers {
		td := td
		t.Run(td.name, func(t *testing.T) {
			t.Parallel()

			for _, filename := range corpusFiles(t) {
				data, err := ioutil.ReadFile(filename)
				require.NoError(t, err)
				_, err = td.fn(data)
				assert.NoErrorf(t, err, "error from file: %s\n with data:\n%s", filename, string(data))

			}

			for _, name := range jsontestsuiteFiles(t) {
				data, err := ioutil.ReadFile(name)
				require.NoError(t, err)
				_, err = td.fn(data)
				assert.NoErrorf(t, err, "error from file: %s\n with data:\n%s", name, string(data))
			}
		})
	}
}

// various values that fuzz has crashed on in the past
var oldCrashers = []string{
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
	"{\"�\":\"0\",\"\xbd\":\"\"}",
}

func jsontestsuiteFiles(t testing.TB) []string {
	t.Helper()
	var files []string
	testdir := filepath.FromSlash("testdata/jsontestsuite")
	dir, err := ioutil.ReadDir(testdir)
	require.NoError(t, err)
	for _, info := range dir {
		name := info.Name()
		if info.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		files = append(files, filepath.Join(testdir, name))
	}
	return files
}

func TestValid(t *testing.T) {
	t.Run("invalidJSON", func(t *testing.T) {
		for _, s := range invalidJSON {
			assert.Equalf(t,
				json.Valid([]byte(s)),
				Valid([]byte(s), nil),
				"failed on input: %s", s,
			)
		}
	})

	for _, name := range jsontestsuiteFiles(t) {
		data, err := ioutil.ReadFile(name)
		require.NoError(t, err)
		t.Run(name, func(t *testing.T) {
			var want bool
			switch filepath.Base(name)[0] {
			case 'y':
				want = true
			case 'n':
				want = false
			default:
				want = json.Valid(data)
			}
			got := Valid(data, nil)
			require.Equal(t, want, got, string(data))
		})
	}

	t.Run("oldCrashers", func(t *testing.T) {
		for _, s := range oldCrashers {
			assert.Equalf(t,
				json.Valid([]byte(s)),
				Valid([]byte(s), nil),
				"failed on input: %s", s,
			)
		}
	})

	for _, s := range jsonTestFiles {
		t.Run(s, func(t *testing.T) {
			data := getTestdataJSONGz(t, s)
			assert.Equal(t, json.Valid(data), Valid(data, nil))
		})
	}
}

func TestSkipValue(t *testing.T) {
	for _, s := range jsonTestFiles {
		t.Run(s, func(t *testing.T) {
			data := getTestdataJSONGz(t, s)
			skippedBytes, err := SkipValue(data, nil)
			wantSkippedBytes, wantErr := skipValueEquiv(data)
			if wantErr != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, wantSkippedBytes, skippedBytes, "not equal for %q", s)
		})
	}

	t.Run("invalid json", func(t *testing.T) {
		for _, s := range invalidJSON {
			data := []byte(s)
			skippedBytes, err := SkipValue(data, nil)
			wantSkippedBytes, wantErr := skipValueEquiv(data)
			if wantErr != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, wantSkippedBytes, skippedBytes, "not equal for %q", s)
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

func getUnmarshalResult(data []byte, v interface{}) (offset int, null bool, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()

	switch token {
	case json.Delim('['), json.Delim('{'):
		decoder = json.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&v)
		return int(decoder.InputOffset()), false, err
	}
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
	t.Helper()
	if got == nil {
		return assertNilMatchesUnmarshal(t, data, gotP, gotErr)
	}
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

func assertNilMatchesUnmarshal(t *testing.T, data []byte, gotP int, gotErr error) int {
	t.Helper()
	var umVal interface{}
	umOffset, _, umErr := getUnmarshalResult(data, &umVal)
	pp := gotP
	if umOffset > pp {
		pp = umOffset
	}
	if umVal == nil {
		if umErr == nil {
			assert.NoError(t, gotErr)
			assert.Equal(t, umOffset, gotP)
			return pp
		}
		assert.Error(t, gotErr)
		return 0
	}
	if umErr != nil {
		assert.Error(t, gotErr)
		return 0
	}
	assert.Failf(t, "wrong val", "expected %v but got nil", umVal)
	return pp
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
			p, err = SkipValueFast(data, buf)
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
	simpleValueHandler ArrayValueHandler
}

func (h *simpleValueHandler) HandleObjectValue(fieldname, data []byte) (int, error) {
	return h.HandleArrayValue(data)
}

func (h *simpleValueHandler) HandleArrayValue(data []byte) (int, error) {
	tknType, p, err := NextTokenType(data)
	if err != nil {
		return p, err
	}
	p--
	data = data[p:]
	var pp int
	switch tknType {
	case ObjectStartType:
		pp, err = HandleObjectValues(data, h, &h.buffer)
	case ArrayStartType:
		pp, err = HandleArrayValues(data, h, &h.buffer)
	default:
		pp, err = h.simpleValueHandler.HandleArrayValue(data)
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
