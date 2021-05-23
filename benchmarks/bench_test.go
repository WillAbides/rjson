package benchmarks

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetRepoValues(b *testing.B) {
	runBenchers(b, func(x bencher) bool {
		_, ok := x.(repoDataReader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(repoDataReader)
		data := getTestdata(b, "benchmark_data/github_repo.json")
		var result repoData
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result = repoData{}
			err = runner.readRepoData(data, &result)
		}
		assert.NoError(b, err)
	})
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
	runBenchers(b, func(x bencher) bool {
		_, ok := x.(validator)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(validator)
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			boolResult = runner.valid(data)
		}
		require.True(b, boolResult)
	})
}

func BenchmarkDistinctUserIDs(b *testing.B) {
	data := getTestdata(b, "benchmark_data/twitter.json")
	var result []int64
	var err error
	runBenchers(b, func(x bencher) bool {
		_, ok := x.(distinctUserIDser)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(distinctUserIDser)
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			result, err = runner.distinctUserIDs(data, result[:0])
			if err != nil {
				break
			}
		}
		require.NoError(b, err)
	})
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
	runBenchers(b, func(x bencher) bool {
		_, ok := x.(objectReader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(objectReader)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			mapResult, err = runner.readObject(data)
		}
		require.NoError(b, err)
		_ = mapResult
	})
}

func BenchmarkReadFloat64_zero(b *testing.B) {
	runBenchReadFloat64(b, []byte(`0`))
}

func BenchmarkReadFloat64_smallInt(b *testing.B) {
	runBenchReadFloat64(b, []byte(`42`))
}

func BenchmarkReadFloat64_negExp(b *testing.B) {
	runBenchReadFloat64(b, []byte(`-42.123e5`))
}

var float64Result float64

func runBenchReadFloat64(b *testing.B, data []byte) {
	runBenchers(b, func(x bencher) bool {
		if x.name() == "encoding_json" {
			return false
		}
		_, ok := x.(float64Reader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(float64Reader)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			float64Result, err = runner.readFloat64(data)
		}
		require.NoError(b, err)
		_ = float64Result
	})
}

func BenchmarkReadInt64_zero(b *testing.B) {
	runBenchReadInt64(b, []byte(`0`))
}

func BenchmarkReadInt64_small(b *testing.B) {
	runBenchReadInt64(b, []byte(`42`))
}

func BenchmarkReadInt64_big_negative(b *testing.B) {
	runBenchReadInt64(b, []byte(`-9223372036854775807`))
}

var int64Result int64

func runBenchReadInt64(b *testing.B, data []byte) {
	runBenchers(b, func(x bencher) bool {
		if x.name() == "encoding_json" {
			return false
		}
		_, ok := x.(int64Reader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(int64Reader)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			int64Result, err = runner.readInt64(data)
		}
		require.NoError(b, err)
		_ = int64Result
	})
}

func BenchmarkDecodeInt64_zero(b *testing.B) {
	runBenchDecodeInt64(b, []byte(`0`))
}

func BenchmarkDecodeInt64_small(b *testing.B) {
	runBenchDecodeInt64(b, []byte(`42`))
}

func BenchmarkDecodeInt64_big_negative(b *testing.B) {
	runBenchDecodeInt64(b, []byte(`-9223372036854775807`))
}

func runBenchDecodeInt64(b *testing.B, data []byte) {
	runBenchers(b, func(x bencher) bool {
		_, ok := x.(int64Decoder)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(int64Decoder)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			err = runner.decodeInt64(data, &int64Result)
		}
		require.NoError(b, err)
		_ = int64Result
	})
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
	runBenchers(b, func(x bencher) bool {
		if x.name() == "encoding_json" {
			return false
		}
		_, ok := x.(stringReader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(stringReader)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			stringResult, err = runner.readString(data)
		}
		require.NoError(b, err)
		_ = stringResult
	})
}

func BenchmarkReadBool(b *testing.B) {
	data := []byte(`true`)
	runBenchers(b, func(x bencher) bool {
		if x.name() == "encoding_json" {
			return false
		}
		_, ok := x.(boolReader)
		return ok
	}, func(b *testing.B, bb bencher) {
		runner := bb.(boolReader)
		var err error
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			boolResult, err = runner.readBool(data)
		}
		require.NoError(b, err)
		_ = boolResult
	})
}

func runBenchers(b *testing.B, filter func(bencher) bool, fn func(b *testing.B, bnchr bencher)) {
	b.Helper()
	bnchrs := getBenchers(filter)
	if len(bnchrs) == 0 {
		b.SkipNow()
		return
	}
	if len(bnchrs) == 1 && os.Getenv("BENCHER") != "" {
		initBencher(bnchrs[0])
		fn(b, bnchrs[0])
		return
	}

	for i := range bnchrs {
		b.Run(bnchrs[i].name(), func(bb *testing.B) {
			initBencher(bnchrs[i])
			fn(bb, bnchrs[i])
		})
	}
}
