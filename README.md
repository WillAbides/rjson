# rjson

[![godoc](https://pkg.go.dev/badge/github.com/willabides/rjson.svg)](https://pkg.go.dev/github.com/willabides/rjson)
[![ci](https://github.com/WillAbides/rjson/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/rjson/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

rjson is a json parser that relies on [Ragel-generated](http://www.colm.net/open-source/ragel/) state machines for most
parsing. rjson's api is minimal and focussed on efficient parsing.

## Benchmarks

### JSON data

With the exception of github_repo.json, this is all taken
from [nativejson-benchmark](https://github.com/miloyip/nativejson-benchmark)

All tested JSON data are in UTF-8.

JSON file   | Size | Description
------------|------|-----------------------
[canada.json](testdata/benchmark_data/canada.json) [source](https://github.com/mloskot/json_benchmark/blob/master/data/canada.json) | 2199KB | Contour of Canada border in [GeoJSON](http://geojson.org/) format. Contains a lot of real numbers.
[citm_catalog.json](testdata/benchmark_data/citm_catalog.json) [source](https://github.com/RichardHightower/json-parsers-benchmark/blob/master/data/citm_catalog.json) | 1737KB | A big benchmark file with indentation used in several Java JSON parser benchmarks.
[twitter.json](testdata/benchmark_data/twitter.json) | 632KB | Search "一" (character of "one" in Japanese and Chinese) in Twitter public time line for gathering some tweets with CJK characters.
[github_repo.json](testdata/benchmark_data/github_repo.json) | 6KB | The golang/go repository from GitHub's API. `curl https://api.github.com/repos/golang/go`

### Compared packages

The go community is fortunate to have multiple high-performance options for parsing JSON. The benchmarks include some
better known packages. They don't all have the same features, so not all the compared packages are represented in
every benchmark.

Every package should be able to put their best foot forward in the benchmarks. I have tried to write the most 
performant code for each package in the benchmark. If anybody sees a way to improve the benchmarks for any package, 
please create an issue to let me know.

This import statement has all the players seen in the benchmarks below:

```go
import (
	encoding_json "encoding/json"

	jsonparser "github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
	gjson "github.com/tidwall/gjson"
	fastjson "github.com/valyala/fastjson"
	rjson "github.com/willabides/rjson"
)
```

### ReadObject()

This measures the ability to create a `map[string]interface{}` from a JSON object. That isn't an operation you want
to do often where performance matters, but ironically it makes a pretty good comparative benchmark because it covers
parsing of several types.

```
BenchmarkReadObject_canada/encoding_json-8           27	 40039902 ns/op   56.22 MB/s  10944057 B/op  392541 allocs/op
BenchmarkReadObject_canada/rjson-8                   44	 24723614 ns/op   91.05 MB/s   7892995 B/op  223659 allocs/op
BenchmarkReadObject_canada/gjson-8                   21	 53121043 ns/op   42.38 MB/s  10551298 B/op  281393 allocs/op
BenchmarkReadObject_canada/jsoniter-8                21	 55269513 ns/op   40.73 MB/s  16188464 B/op  665992 allocs/op

BenchmarkReadObject_citm_catalog/encoding_json-8     72	 17082554 ns/op  101.11 MB/s   5123046 B/op   95395 allocs/op
BenchmarkReadObject_citm_catalog/rjson-8            130	  9163264 ns/op  188.49 MB/s   4899119 B/op   76435 allocs/op
BenchmarkReadObject_citm_catalog/gjson-8             74	 15865039 ns/op  108.87 MB/s   6433901 B/op   54373 allocs/op
BenchmarkReadObject_citm_catalog/jsoniter-8         100	 10602162 ns/op  162.91 MB/s   5571741 B/op  118767 allocs/op

BenchmarkReadObject_github_repo/encoding_json-8   13154	    90680 ns/op   67.66 MB/s     25695 B/op     503 allocs/op
BenchmarkReadObject_github_repo/rjson-8           40250	    29772 ns/op  206.07 MB/s     17046 B/op     318 allocs/op
BenchmarkReadObject_github_repo/gjson-8           37813	    31576 ns/op  194.29 MB/s     22676 B/op     122 allocs/op
BenchmarkReadObject_github_repo/jsoniter-8        28158	    42690 ns/op  143.71 MB/s     25461 B/op     531 allocs/op

BenchmarkReadObject_twitter/encoding_json-8         164	  7285272 ns/op   86.68 MB/s   2152334 B/op   31267 allocs/op
BenchmarkReadObject_twitter/rjson-8                 321	  3755500 ns/op  168.16 MB/s   2969081 B/op   28473 allocs/op
BenchmarkReadObject_twitter/gjson-8                 187	  6347622 ns/op   99.49 MB/s   2396039 B/op   12183 allocs/op
BenchmarkReadObject_twitter/jsoniter-8              236	  5082547 ns/op  124.25 MB/s   2426520 B/op   45043 allocs/op
```

### Repo Values

This tests getting three values of different types from github_repo.json.

```
BenchmarkGetRepoValues/encoding_json-8     30532    37379 ns/op   248 B/op	  7 allocs/op
BenchmarkGetRepoValues/rjson-8            588814     1977 ns/op    16 B/op	  1 allocs/op
BenchmarkGetRepoValues/gjson-8            524803     2247 ns/op   200 B/op	  4 allocs/op
BenchmarkGetRepoValues/jsoniter-8         493632     2367 ns/op   360 B/op	 25 allocs/op
BenchmarkGetRepoValues/jsonparser-8       506372     2262 ns/op    16 B/op	  1 allocs/op
BenchmarkGetRepoValues/fastjson-8         246114     4553 ns/op    16 B/op	  1 allocs/op
```

### Validating JSON

This benchmark is for each package's equivalent of running `encoding_json.Valid()`.

```
BenchmarkValid_canada/encoding_json-8           178	 6445099 ns/op   349.27 MB/s       6 B/op      0 allocs/op
BenchmarkValid_canada/rjson-8                   543	 2191929 ns/op  1026.97 MB/s       0 B/op      0 allocs/op
BenchmarkValid_canada/gjson-8                   510	 2358083 ns/op   954.61 MB/s       0 B/op      0 allocs/op
BenchmarkValid_canada/jsoniter-8                259	 4622866 ns/op   486.94 MB/s      67 B/op      8 allocs/op
BenchmarkValid_canada/fastjson-8                388	 3071372 ns/op   732.91 MB/s       0 B/op      0 allocs/op

BenchmarkValid_citm_catalog/encoding_json-8     237	 5056105 ns/op   341.61 MB/s       5 B/op      0 allocs/op
BenchmarkValid_citm_catalog/rjson-8             748	 1530737 ns/op  1128.35 MB/s       0 B/op      0 allocs/op
BenchmarkValid_citm_catalog/gjson-8             760	 1547381 ns/op  1116.21 MB/s       0 B/op      0 allocs/op
BenchmarkValid_citm_catalog/jsoniter-8          423	 2830790 ns/op   610.15 MB/s  240658 B/op  25874 allocs/op
BenchmarkValid_citm_catalog/fastjson-8          728	 1570642 ns/op  1099.68 MB/s       0 B/op      0 allocs/op

BenchmarkValid_github_repo/encoding_json-8    56817	   19879 ns/op   308.62 MB/s       0 B/op      0 allocs/op
BenchmarkValid_github_repo/rjson-8           217376	    5173 ns/op  1185.98 MB/s       0 B/op      0 allocs/op
BenchmarkValid_github_repo/gjson-8           147892	    7428 ns/op   825.97 MB/s       0 B/op      0 allocs/op
BenchmarkValid_github_repo/jsoniter-8        102781	   11120 ns/op   551.72 MB/s    1576 B/op    122 allocs/op
BenchmarkValid_github_repo/fastjson-8        164803	    6950 ns/op   882.75 MB/s       0 B/op      0 allocs/op

BenchmarkValid_twitter/encoding_json-8          556	 2112273 ns/op   298.97 MB/s       2 B/op      0 allocs/op
BenchmarkValid_twitter/rjson-8                 1951	  573490 ns/op  1101.18 MB/s       0 B/op      0 allocs/op
BenchmarkValid_twitter/gjson-8                 1518	  748273 ns/op   843.96 MB/s       0 B/op      0 allocs/op
BenchmarkValid_twitter/jsoniter-8               751	 1573439 ns/op   401.36 MB/s  375968 B/op  15274 allocs/op
BenchmarkValid_twitter/fastjson-8              1449	  780132 ns/op   809.50 MB/s       0 B/op      0 allocs/op
```
