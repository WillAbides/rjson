package benchmarks

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func getTestdata(t testing.TB, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(filepath.Join("..", "testdata", filepath.FromSlash(path)))
	require.NoError(t, err)
	return b
}
