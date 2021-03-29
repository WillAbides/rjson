package benchmarks

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkReadObject_canada(b *testing.B) {
	benchReadObject(b, getTestdata(b, "benchmark_data/canada.json"))
}

func BenchmarkReadObject_citm_catalog(b *testing.B) {
	benchReadObject(b, getTestdata(b, "benchmark_data/citm_catalog.json"))
}

func BenchmarkReadObject_github_repo(b *testing.B) {
	benchReadObject(b, getTestdata(b, "benchmark_data/github_repo.json"))
}

func BenchmarkReadObject_twitter(b *testing.B) {
	benchReadObject(b, getTestdata(b, "benchmark_data/twitter.json"))
}

var readObjectResult map[string]interface{}

func benchReadObject(b *testing.B, data []byte) {
	for _, runner := range readObjectRunners {
		var pool sync.Pool
		var err error
		b.Run(runner.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				readObjectResult, err = runner.fn(data, &pool)
			}
			require.NoError(b, err)
			_ = readObjectResult
		})
	}
}

func Test_readObjectRunners(t *testing.T) {
	var golden readObjectRunner
	for _, runner := range readObjectRunners {
		if runner.name == "encoding_json" {
			golden = runner
			break
		}
	}

	for _, testFile := range []string{
		"benchmark_data/canada.json",
		"benchmark_data/citm_catalog.json",
		"benchmark_data/github_repo.json",
		"benchmark_data/twitter.json",
	} {
		t.Run(testFile, func(t *testing.T) {
			origData := getTestdata(t, testFile)
			want, wantErr := golden.fn(origData, &sync.Pool{})
			for _, runner := range readObjectRunners {
				t.Run(runner.name, func(t *testing.T) {
					var pool sync.Pool
					data := make([]byte, len(origData))
					copy(data, origData)
					got, err := runner.fn(data, &pool)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, got)
					} else {
						assert.Error(t, err)
					}

					// do it again with the same pool
					data = make([]byte, len(origData))
					copy(data, origData)
					got, err = runner.fn(data, &pool)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, got)
					} else {
						assert.Error(t, err)
					}
				})
			}
		})
	}
}
