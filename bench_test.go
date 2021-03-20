package rjson

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

var (
	benchInt  int
	benchBool bool
)

func BenchmarkCompareSkip(b *testing.B) {
	for _, file := range jsonTestFiles {
		b.Run(file, func(b *testing.B) {
			data := getTestdataJSONGz(b, file)
			size := int64(len(data))
			var err error
			b.Run("rjson", func(b *testing.B) {
				buf := &Buffer{}
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

func BenchmarkSkip(b *testing.B) {
	var err error
	for _, file := range []string{"citm_catalog.json"} {
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
