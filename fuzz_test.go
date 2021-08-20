// +build gofuzzbeta

package rjson

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzMe(f *testing.F) {
	corpusDir := filepath.FromSlash(`testdata/fuzz/corpus`)
	dir, err := ioutil.ReadDir(corpusDir)
	require.NoError(f, err)
	for _, info := range dir {
		if info.IsDir() {
			continue
		}
		dd, err := ioutil.ReadFile(filepath.Join(corpusDir, info.Name()))
		require.NoError(f, err)
		f.Add(string(dd))
	}

	f.Fuzz(func(t *testing.T, data string) {
		for _, fd := range fuzzers {
			d := make([]byte, len(data))
			copy(d, data)
			_, err := fd.fn(d)
			require.NoError(t, err)
		}
	})
}
