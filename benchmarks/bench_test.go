package benchmarks

import (
	"fmt"
	"strings"
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

var float64Result float64

func runBenchFloat64(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(float64Reader)
		if !ok {
			continue
		}
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				float64Result, err = runner.readFloat64(data)
			}
			require.NoError(b, err)
			_ = float64Result
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

var int64Result int64

func runBenchReadInt64(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(int64Reader)
		if !ok {
			continue
		}
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				int64Result, err = runner.readInt64(data)
			}
			require.NoError(b, err)
			_ = int64Result
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

var mapResult map[string]interface{}

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
				mapResult, err = runner.readObject(data)
			}
			require.NoError(b, err)
			_ = mapResult
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

var boolResult bool

func benchValid(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(validator)
		if !ok {
			continue
		}
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				boolResult = runner.valid(data)
			}
			require.True(b, boolResult)
		})
	}
}

func BenchmarkReadString_short_ascii(b *testing.B) {
	benchReadString(b, []byte(`"hello"`))
}

func BenchmarkReadString_medium_ascii(b *testing.B) {
	benchReadString(b, []byte(fmt.Sprintf(`%q`, strings.Repeat("All work and no play makes Jack a dull boy.\n", 20))))
}

func BenchmarkReadString_medium(b *testing.B) {
	benchReadString(b, []byte(`"@aym0566x \n\n名前:前田あゆみ\n第一印象:なんか怖っ！\n今の印象:とりあえずキモい。噛み合わない\n好きなところ:ぶすでキモいとこ😋✨✨\n思い出:んーーー、ありすぎ😊❤️\nLINE交換できる？:あぁ……ごめん✋\nトプ画をみて:照れますがな😘✨\n一言:お前は一生もんのダチ💖"`))
}

var stringResult string

func benchReadString(b *testing.B, data []byte) {
	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(stringReader)
		if !ok {
			continue
		}
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				stringResult, err = runner.readString(data)
			}
			require.NoError(b, err)
			_ = stringResult
		})
	}
}

func BenchmarkReadBool(b *testing.B) {
	for _, bb := range benchers {
		data := []byte(`true`)
		initBencher(bb)
		runner, ok := bb.(boolReader)
		if !ok {
			continue
		}
		var err error
		b.Run(bb.name(), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				boolResult, err = runner.readBool(data)
			}
			require.NoError(b, err)
			_ = boolResult
		})
	}
}
