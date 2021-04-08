package rjson

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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

func testFuzzerFunc(t *testing.T, fn func([]byte) (int, error)) {
	t.Helper()
	require.NotNil(t, fn)
	for _, filename := range corpusFiles(t) {
		data, err := ioutil.ReadFile(filename)
		require.NoError(t, err)
		_, err = fn(data)
		assert.NoErrorf(t, err, "error from file: %s\n with data:\n%s", filename, string(data))
	}
}

//nolint:unused,deadcode // this is a convenience method for ad-hoc testing
func testFuzzerWithInput(t *testing.T, fuzzer func([]byte) (int, error), inputs ...string) {
	t.Helper()

	for _, input := range inputs {
		data := []byte(input)
		_, err := fuzzer(data)
		assert.NoError(t, err, "error on input: %s", input)
	}
}

func Test_fuzzJSONValueParseJSON(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzJSONValueParseJSON)
}

func Test_fuzzReadFloat64(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadFloat64)
}

func Test_fuzzReadUint64(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadUint64)
}

func Test_fuzzReadUint32(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadUint32)
}

func Test_fuzzReadUint(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadUint)
}

func Test_fuzzReadInt64(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadInt64)
}

func Test_fuzzReadInt32(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadInt32)
}

func Test_fuzzReadInt(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadInt)
}

func Test_fuzzReadString(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadString)
}

func Test_fuzzReadStringBytes(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadStringBytes)
}

func Test_fuzzReadBool(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadBool)
}

func Test_fuzzReadNull(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadNull)
}

func Test_fuzzSkipValue(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzSkipValue)
}

func Test_fuzzValid(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzValid)
}

func Test_fuzzNextToken(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzNextToken)
}

func Test_adhoc(t *testing.T) {
	testFuzzerWithInput(t, fuzzNextToken, `	
			`)
}

func Test_fuzzNextTokenType(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzNextTokenType)
}

func Test_fuzzHandleArrayValues(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzHandleArrayValues)
}

func Test_fuzzHandleObjectValues(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzHandleObjectValues)
}

func Test_fuzzReadArray(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadArray)
}

func Test_fuzzReadObject(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadObject)
}

func Test_fuzzReadValue(t *testing.T) {
	t.Parallel()
	testFuzzerFunc(t, fuzzReadValue)
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
	"[[0e],0e0]",
	"[0.,0e0]",
	"[0E,0E0]",
	"[0.,0.0]",
	"[0e,0e0]",
	"{\"\":0.,\"\":0.0}",
	"[0E,0.0]",
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
