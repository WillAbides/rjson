package rjson

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var benchInt int

func BenchmarkSkip(b *testing.B) {
	var err error
	for _, file := range jsonTestFiles {
		buf := &Buffer{}
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))
		b.Run(file, func(b *testing.B) {
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
}
