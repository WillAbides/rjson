package benchmarks

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkValid_canada(b *testing.B) {
	benchValid(b, getTestdata(b, "benchmark_data/canada.json"))
}

func BenchmarkValid_citm_catalog(b *testing.B) {
	benchValid(b, getTestdata(b, "benchmark_data/citm_catalog.json"))
}

func BenchmarkValid_github_repo(b *testing.B) {
	benchValid(b, getTestdata(b, "benchmark_data/github_repo.json"))
}

func BenchmarkValid_twitter(b *testing.B) {
	benchValid(b, getTestdata(b, "benchmark_data/twitter.json"))
}

func benchValid(b *testing.B, data []byte) {
	for _, runner := range validRunners {
		var pool sync.Pool
		var result bool
		b.Run(runner.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				result = runner.fn(data, &pool)
			}
			require.True(b, result)
		})
	}
}

func Test_validRunners(t *testing.T) {
	testdir := filepath.FromSlash("../testdata/jsontestsuite")
	dir, err := ioutil.ReadDir(testdir)
	require.NoError(t, err)

	// these two tests cause fastjson to hang
	fastjsonSkips := map[string]bool{
		`n_structure_100000_opening_arrays.json`: true,
		`n_structure_open_array_object.json`:     true,
	}

	for _, runner := range validRunners {
		t.Run(runner.name, func(t *testing.T) {
			if runner.name == "jsoniter" {
				t.Skip(`This is a reported issue. Remove this skip when https://github.com/json-iterator/go/issues/540 is addressed`)
			}
			for _, fileInfo := range dir {
				name := fileInfo.Name()
				if fileInfo.IsDir() || !strings.HasSuffix(name, ".json") || name == "" {
					continue
				}
				if runner.name == `fastjson` && fastjsonSkips[name] {
					continue
				}
				var want bool
				switch name[0] {
				case 'y':
					want = true
				case 'n':
					want = false
				default:
					continue
				}
				var pool sync.Pool
				origData, err := ioutil.ReadFile(filepath.Join(filepath.FromSlash("../testdata/jsontestsuite"), name))
				require.NoError(t, err)
				t.Run(name, func(t *testing.T) {
					data := make([]byte, len(origData))
					copy(data, origData)
					got := runner.fn(data, &pool)
					assert.Equalf(t, want, got, "data: %s", string(data))
					assert.Equal(t, origData, data)

					// do it again with the same pool
					data = make([]byte, len(origData))
					copy(data, origData)
					got = runner.fn(data, &pool)
					assert.Equalf(t, want, got, "data: %s", string(data))
					assert.Equal(t, origData, data)
				})
			}
		})
	}
}
