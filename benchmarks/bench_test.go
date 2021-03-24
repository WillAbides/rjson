package benchmarks

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/willabides/rjson"
)

func BenchmarkGetValuesFromObject(b *testing.B) {
	type resType struct {
		PublicGists int64  `json:"public_gists"`
		PublicRepos int64  `json:"public_repos"`
		Login       string `json:"login"`
	}

	wantRes := resType{
		PublicGists: 8,
		PublicRepos: 8,
		Login:       "octocat",
	}

	data := getTestdataJSONGz(b, "github_user.json")

	b.Run("rjson", func(b *testing.B) {
		var res resType
		doneErr := fmt.Errorf("done")
		var err error
		buffer := &rjson.Buffer{}
		var stringBuf []byte
		var seenRepos, seenGists, seenLogin bool
		handler := rjson.ObjectValueHandlerFunc(func(fieldname, data []byte) (p int, err error) {
			switch string(fieldname) {
			case `public_gists`:
				res.PublicGists, p, err = rjson.ReadInt64(data)
				seenGists = true
			case `public_repos`:
				res.PublicRepos, p, err = rjson.ReadInt64(data)
				seenRepos = true
			case `login`:
				stringBuf, p, err = rjson.ReadStringBytes(data, stringBuf[:0])
				res.Login = string(stringBuf)
				seenLogin = true
			default:
				p, err = buffer.SkipValue(data)
			}
			if err == nil && seenGists && seenRepos && seenLogin {
				return p, doneErr
			}
			return p, err
		})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			seenGists, seenGists, seenLogin = false, false, false
			_, err = buffer.HandleObjectValues(data, handler)
		}
		require.Equal(b, wantRes, res)
		require.EqualError(b, err, "done")
	})

	b.Run("gjson", func(b *testing.B) {
		var res resType
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := gjson.GetManyBytes(data, `public_gists`, `public_repos`, `login`)
			res.PublicGists = results[0].Int()
			res.PublicRepos = results[1].Int()
			res.Login = results[2].Str
		}
		require.Equal(b, wantRes, res)
	})

	b.Run("encoding/json", func(b *testing.B) {
		var res resType
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = json.Unmarshal(data, &res)
		}
		require.NoError(b, err)
		require.Equal(b, wantRes, res)
	})

	b.Run("jsoniter", func(b *testing.B) {
		var res resType
		iter := jsoniter.NewIterator(jsoniter.ConfigFastest)
		var seenRepos, seenGists, seenLogin bool
		callback := func(it *jsoniter.Iterator, field string) bool {
			switch field {
			case `public_gists`:
				res.PublicGists = it.ReadInt64()
				seenGists = true
			case `public_repos`:
				res.PublicRepos = it.ReadInt64()
				seenRepos = true
			case `login`:
				res.Login = it.ReadString()
				seenLogin = true
			default:
				it.Skip()
			}
			if seenRepos && seenGists && seenLogin {
				return false
			}
			return true
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			seenRepos, seenLogin, seenGists = false, false, false
			iter.ResetBytes(data)
			iter.ReadObjectCB(callback)
		}
		require.Equal(b, wantRes, res)
		require.NoError(b, iter.Error)
	})

	b.Run("jsonparser", func(b *testing.B) {
		var res resType
		doneErr := fmt.Errorf("done")
		var seenRepos, seenGists, seenLogin bool
		callback := func(key, value []byte, dataType jsonparser.ValueType, offset int) error {
			var err error
			switch string(key) {
			case "public_gists":
				res.PublicGists, err = jsonparser.ParseInt(value)
				seenGists = true
			case "public_repos":
				res.PublicRepos, err = jsonparser.ParseInt(value)
				seenRepos = true
			case "login":
				res.Login, err = jsonparser.ParseString(value)
				seenLogin = true
			}
			if err == nil && seenGists && seenRepos && seenLogin {
				return doneErr
			}
			return err
		}
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			seenGists, seenGists, seenLogin = false, false, false
			err = jsonparser.ObjectEach(data, callback)
		}
		require.Equal(b, wantRes, res)
		require.EqualError(b, err, "done")
	})
}

func BenchmarkSkipFast(b *testing.B) {
	for _, sample := range benchSamples(b) {
		data := sample.data
		b.Run(sample.name, func(b *testing.B) {
			b.Run("rjson", func(b *testing.B) {
				var err error
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchInt, err = benchBuf.SkipValueFast(data)
				}
				require.NoError(b, err)
			})

			b.Run("jsonparser", func(b *testing.B) {
				var err error
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					_, _, benchInt, err = jsonparser.Get(data)
				}
				require.NoError(b, err)
			})
		})

	}
}

func BenchmarkValid(b *testing.B) {
	for _, sample := range benchSamples(b) {
		data := sample.data
		b.Run(sample.name, func(b *testing.B) {
			b.Run("rjson", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchBool = benchBuf.Valid(data)
				}
				require.True(b, benchBool)
			})

			b.Run("gjson", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchBool = gjson.ValidBytes(data)
				}
				require.True(b, benchBool)
			})

			b.Run("encoding/json", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchBool = json.Valid(data)
				}
				require.True(b, benchBool)
			})

			b.Run("jsoniter", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchBool = jsoniter.Valid(data)
				}
				require.True(b, benchBool)
			})
		})
	}
}

var (
	benchInt  int
	benchBool bool
	benchBuf  = &rjson.Buffer{}
)

type benchSample struct {
	name string
	data []byte
}

func benchSamples(t testing.TB) []benchSample {
	return []benchSample{
		{
			name: "large object",
			data: getTestdataJSONGz(t, "citm_catalog.json"),
		},
		{
			name: "unicode-heavy object",
			data: getTestdataJSONGz(t, "sample.json"),
		},
		{
			name: "canada.json",
			data: getTestdataJSONGz(t, "canada.json"),
		},
		{
			name: "github user",
			data: getTestdataJSONGz(t, "github_user.json"),
		},
		{
			name: "github repo",
			data: getTestdataJSONGz(t, "github_repo.json"),
		},
		{
			name: "jsonchecker_pass1",
			data: getTestdataJSONGz(t, "jsonchecker_pass1.json"),
		},
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
	src := filepath.Join("..", "testdata", filename+".gz")
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
