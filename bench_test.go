package rjson

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	benchInt  int
	benchBool bool
	benchBuf  = &Buffer{}
)

func BenchmarkSkip(b *testing.B) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))

		b.Run(file, func(b *testing.B) {
			b.Run("SkipValue", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(size)
				var err error
				for i := 0; i < b.N; i++ {
					benchInt, err = benchBuf.SkipValue(data)
				}
				require.NoError(b, err)
			})

			b.Run("SkipValueFast", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(size)
				var err error
				for i := 0; i < b.N; i++ {
					benchInt, err = benchBuf.SkipValueFast(data)
				}
				require.NoError(b, err)
			})
		})
	}
}

func BenchmarkValid(b *testing.B) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))
		b.Run(file, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(size)
			for i := 0; i < b.N; i++ {
				benchBool = benchBuf.Valid(data)
			}
			require.True(b, benchBool)
		})
	}
}

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
	var res resType
	doneErr := fmt.Errorf("done")
	var err error
	buffer := &Buffer{}
	var stringBuf []byte
	var seenRepos, seenGists, seenLogin bool
	handler := ObjectValueHandlerFunc(func(fieldname, data []byte) (p int, err error) {
		switch string(fieldname) {
		case `public_gists`:
			res.PublicGists, p, err = ReadInt64(data)
			seenGists = true
		case `public_repos`:
			res.PublicRepos, p, err = ReadInt64(data)
			seenRepos = true
		case `login`:
			stringBuf, p, err = ReadStringBytes(data, stringBuf[:0])
			res.Login = string(stringBuf)
			seenLogin = true
		default:
			p, err = buffer.SkipValueFast(data)
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
}
