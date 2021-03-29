package benchmarks

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkGetRepoValues(b *testing.B) {
	wantRes := repoValues{
		FullName: "golang/go",
		Forks:    12162,
		Archived: false,
	}
	for _, x := range getRepoValuesRunners {
		data := getTestdata(b, "benchmark_data/github_repo.json")
		var pool sync.Pool
		var result repoValues
		var err error
		fn := x.fn
		b.Run(x.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				result = repoValues{}
				err = fn(data, &pool, &result)
				if err != nil {
					break
				}
			}
			assert.NoError(b, err)
			assert.Equal(b, wantRes, result)
		})
	}
}

func Test_getRepoValuesRunners(t *testing.T) {
	var golden getRepoValuesRunner
	for _, runner := range getRepoValuesRunners {
		if runner.name == "encoding_json" {
			golden = runner
			break
		}
	}
	for _, td := range []struct {
		name string
		data string
	}{
		{
			name: "standard data",
			data: string(getTestdata(t, "benchmark_data/github_repo.json")),
		},
		{
			name: "nulls",
			data: `{"archived": null, "forks": null, "full_name": null}`,
		},
		{
			name: "wrong types",
			data: `{"archived": 12, "forks": "hi", "full_name": true}`,
		},
		{
			name: "empty object",
			data: `{}`,
		},
		{
			name: "archived true at the end",
			data: `{"forks": 1, "full_name": "puppy", "archived": true}`,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			origData := []byte(td.data)
			var want repoValues
			wantErr := golden.fn(origData, &sync.Pool{}, &want)
			for _, runner := range getRepoValuesRunners {
				t.Run(runner.name, func(t *testing.T) {
					var pool sync.Pool
					var result repoValues
					data := make([]byte, len(origData))
					copy(data, origData)
					err := runner.fn(data, &pool, &result)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, result)
					} else {
						assert.Error(t, err)
					}
					assert.Equal(t, origData, data)

					// do it again with the same pool
					result = repoValues{}
					data = make([]byte, len(origData))
					copy(data, origData)
					err = runner.fn(data, &pool, &result)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, result)
					} else {
						assert.Error(t, err)
					}
					assert.Equal(t, origData, data)
				})
			}
		})
	}
}
