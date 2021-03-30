package benchmarks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetRepoValues(b *testing.B) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(repoDataReader)
		if !ok {
			continue
		}
		data := getTestdata(b, "benchmark_data/github_repo.json")
		var result repoData
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				result = repoData{}
				err = runner.readRepoData(data, &result)
			}
			assert.NoError(b, err)
		})
	}
}

func BenchmarkReadFloat_zero(b *testing.B) {
	runBenchFloat64(b, []byte(`0`))
}

func BenchmarkReadFloat64_smallInt(b *testing.B) {
	runBenchFloat64(b, []byte(`42`))
}

func BenchmarkReadFloat64_negExp(b *testing.B) {
	runBenchFloat64(b, []byte(`-42.123e5`))
}

func runBenchFloat64(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(float64Reader)
		if !ok {
			continue
		}
		var err error
		var val float64
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				val, err = runner.readFloat64(data)
			}
			require.NoError(b, err)
			_ = val
		})
	}
}

func BenchmarkReadInt64_zero(b *testing.B) {
	runBenchReadInt64(b, []byte(`0`))
}

func BenchmarkReadInt64_small(b *testing.B) {
	runBenchReadInt64(b, []byte(`42`))
}

func BenchmarkReadInt64_big_negative(b *testing.B) {
	runBenchReadInt64(b, []byte(`9223372036854775807`))
}

func runBenchReadInt64(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(int64Reader)
		if !ok {
			continue
		}
		var err error
		var val int64
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				val, err = runner.readInt64(data)
			}
			require.NoError(b, err)
			_ = val
		})
	}
}

func BenchmarkReadObject_canada(b *testing.B) {
	runBenchReadObject(b, getTestdata(b, "benchmark_data/canada.json"))
}

func BenchmarkReadObject_citm_catalog(b *testing.B) {
	runBenchReadObject(b, getTestdata(b, "benchmark_data/citm_catalog.json"))
}

func BenchmarkReadObject_github_repo(b *testing.B) {
	runBenchReadObject(b, getTestdata(b, "benchmark_data/github_repo.json"))
}

func BenchmarkReadObject_twitter(b *testing.B) {
	runBenchReadObject(b, getTestdata(b, "benchmark_data/twitter.json"))
}

var readObjectResult map[string]interface{}

func runBenchReadObject(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(objectReader)
		if !ok {
			continue
		}
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				readObjectResult, err = runner.readObject(data)
			}
			require.NoError(b, err)
			_ = readObjectResult
		})
	}
}

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
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(validator)
		if !ok {
			continue
		}
		var result bool
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				result = runner.valid(data)
			}
			require.True(b, result)
		})
	}
}
