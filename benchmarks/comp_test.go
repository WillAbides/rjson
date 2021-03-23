// +build comp

package benchmarks

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

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

	data := []byte(exampleGithubUser)

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
			if err == nil && seenGists && seenRepos && seenLogin{
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
}

func BenchmarkSkipValue(b *testing.B) {
	for _, sample := range benchSamples(b) {
		data := sample.data
		b.Run(sample.name, func(b *testing.B) {
			b.Run("rjson", func(b *testing.B) {
				buffer := &rjson.Buffer{}
				var err error
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					benchInt, err = buffer.SkipValue(data)
				}
				require.NoError(b, err)
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
				iter := jsoniter.NewIterator(jsoniter.ConfigFastest)
				var err error
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				for i := 0; i < b.N; i++ {
					iter.ResetBytes(data)
					iter.Skip()
					err = iter.Error
				}
				if err == io.EOF {
					err = nil
				}
				require.NoError(b, err)
			})
		})
	}
}
