package benchmarks

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestdata(t testing.TB, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(filepath.Join("..", "testdata", filepath.FromSlash(path)))
	require.NoError(t, err)
	return b
}

func Test_objectReaders(t *testing.T) {
	golden := &jsonBencher{}
	for _, testFile := range []string{
		"benchmark_data/canada.json",
		"benchmark_data/citm_catalog.json",
		"benchmark_data/github_repo.json",
		"benchmark_data/twitter.json",
	} {
		t.Run(testFile, func(t *testing.T) {
			origData := getTestdata(t, testFile)
			want, wantErr := golden.readObject(origData)

			for _, bb := range benchers {
				initBencher(bb)
				runner, ok := bb.(objectReader)
				if !ok {
					continue
				}
				t.Run(bb.name(), func(t *testing.T) {
					data := make([]byte, len(origData))
					copy(data, origData)
					got, err := runner.readObject(data)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, got)
					} else {
						assert.Error(t, err)
					}
					assert.Equal(t, origData, data)

					// do it again
					data = make([]byte, len(origData))
					copy(data, origData)
					got, err = runner.readObject(data)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, got)
					} else {
						assert.Error(t, err)
					}
					assert.Equal(t, origData, data)
				})
			}
		})
	}
}

func Test_getRepoValuesRunners(t *testing.T) {
	golden := &jsonBencher{}

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
			var want repoData
			wantErr := golden.readRepoData(origData, &want)

			for _, bb := range benchers {
				initBencher(bb)
				runner, ok := bb.(repoDataReader)
				if !ok {
					continue
				}
				t.Run(bb.name(), func(t *testing.T) {
					var result repoData
					data := make([]byte, len(origData))
					copy(data, origData)
					err := runner.readRepoData(data, &result)
					if wantErr == nil {
						assert.NoError(t, err)
						assert.Equal(t, want, result)
					} else {
						assert.Error(t, err)
					}
					assert.Equal(t, origData, data)

					// do it again
					data = make([]byte, len(origData))
					copy(data, origData)
					err = runner.readRepoData(data, &result)
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

func TestReadInt64(t *testing.T) {
	readTests := []struct {
		data    string
		want    int64
		wantErr bool
	}{
		{data: `42 `, want: 42},
		{data: ` 42`, want: 42},
		{data: `42.1`, wantErr: true},
		{data: `9223372036854775807`, want: 9223372036854775807},
		{data: `92233720368547758070`, wantErr: true},
		{data: `9223372036854775808`, wantErr: true},
		{data: `-9223372036854775808`, want: -9223372036854775808},
		{data: `-9223372036854775809`, wantErr: true},
	}

	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(int64Reader)
		if !ok {
			continue
		}
		t.Run(bb.name(), func(t *testing.T) {
			for _, td := range readTests {
				t.Run(td.data, func(t *testing.T) {
					got, err := runner.readInt64([]byte(td.data))
					if td.wantErr {
						assert.Error(t, err)
						return
					}
					assert.NoError(t, err)
					assert.Equal(t, td.want, got)
				})
			}
		})
	}
}

func TestReadFloat64(t *testing.T) {
	readTests := []struct {
		data    string
		want    float64
		wantErr bool
	}{
		{data: `42 `, want: 42},
		{data: ` 42`, want: 42},
		{data: `42.1`, want: 42.1},
		{data: `9223372036854775807`, want: 9223372036854775807},
		{data: `-42.123e5`, want: -42.123e5},
		{data: `-9223372036854775808`, want: -9223372036854775808},
	}

	for _, bb := range benchers {
		initBencher(bb)
		runner, ok := bb.(float64Reader)
		if !ok {
			continue
		}
		t.Run(bb.name(), func(t *testing.T) {
			for _, td := range readTests {
				t.Run(td.data, func(t *testing.T) {
					got, err := runner.readFloat64([]byte(td.data))
					if td.wantErr {
						assert.Error(t, err)
						return
					}
					assert.NoError(t, err)
					assert.Equal(t, td.want, got)
				})
			}
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

	for _, bb := range benchers {
		runner, ok := bb.(validator)
		if !ok {
			continue
		}
		runnerName := bb.name()
		t.Run(runnerName, func(t *testing.T) {
			initBencher(runner)

			if runnerName == "jsoniter" {
				t.Skip(`This is a reported issue. Remove this skip when https://github.com/json-iterator/go/issues/540 is addressed`)
			}
			for _, fileInfo := range dir {
				name := fileInfo.Name()
				if fileInfo.IsDir() || !strings.HasSuffix(name, ".json") || name == "" {
					continue
				}
				if runnerName == `fastjson` && fastjsonSkips[name] {
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
				origData, err := ioutil.ReadFile(filepath.Join(filepath.FromSlash("../testdata/jsontestsuite"), name))
				require.NoError(t, err)

				t.Run(name, func(t *testing.T) {
					data := make([]byte, len(origData))
					copy(data, origData)
					got := runner.valid(data)
					assert.Equalf(t, want, got, "data: %s", string(data))
					assert.Equal(t, origData, data)

					// do it again with the same pool
					data = make([]byte, len(origData))
					copy(data, origData)
					got = runner.valid(data)
					assert.Equalf(t, want, got, "data: %s", string(data))
					assert.Equal(t, origData, data)
				})
			}
		})
	}
}
