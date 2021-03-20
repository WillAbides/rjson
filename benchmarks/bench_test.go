package benchmarks

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/willabides/rjson"
)

var (
	benchInt  int
	benchBool bool
)

func BenchmarkCompareSkip(b *testing.B) {
	for _, file := range jsonTestFiles {
		// for _, file := range []string{"canada.json"} {
		b.Run(file, func(b *testing.B) {
			data := getTestdataJSONGz(b, file)
			size := int64(len(data))
			var err error
			b.Run("rjson", func(b *testing.B) {
				buf := &rjson.Buffer{}
				b.ReportAllocs()
				b.SetBytes(size)
				for i := 0; i < b.N; i++ {
					benchInt, err = buf.SkipValue(data)
					if err != nil {
						break
					}

				}
				require.NoError(b, err)
			})

			b.Run("jsoniter", func(b *testing.B) {
				b.Skip()
				b.ReportAllocs()
				b.SetBytes(size)
				iter := jsoniter.NewIterator(jsoniter.ConfigFastest)
				for i := 0; i < b.N; i++ {
					iter.ResetBytes(data)
					iter.Skip()
				}
				require.NoError(b, iter.Error)
			})

			b.Run("gjson", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(size)
				for i := 0; i < b.N; i++ {
					benchBool = gjson.ValidBytes(data)
					if !benchBool {
						break
					}
				}
				require.True(b, benchBool)
			})

			b.Run("json.Valid", func(b *testing.B) {
				b.Skip()
				b.ReportAllocs()
				b.SetBytes(size)
				for i := 0; i < b.N; i++ {
					benchBool = json.Valid(data)
					if !benchBool {
						break
					}
				}
				require.True(b, benchBool)
			})
		})
	}
}

func gunzipTestJSON(t testing.TB, filename string) string {
	t.Helper()
	targetDir := filepath.Join("..", "testdata", "tmp")
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
