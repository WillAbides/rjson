package rjson

import (
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
