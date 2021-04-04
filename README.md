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
BenchmarkReadObject_canada/encoding_json-8            26	39743078 ns/op	 56.64 MB/s 10943959 B/op   392541 allocs/op
BenchmarkReadObject_canada/rjson-8                    43	24911271 ns/op	 90.36 MB/s  7892957 B/op   223658 allocs/op
BenchmarkReadObject_canada/gjson-8                    21	52547642 ns/op	 42.84 MB/s 10551293 B/op   281393 allocs/op
BenchmarkReadObject_canada/jsoniter-8                 20	56095794 ns/op	 40.13 MB/s 16188466 B/op   665992 allocs/op

BenchmarkReadObject_citm_catalog/encoding_json-8      66	17171046 ns/op	100.59 MB/s  5123188 B/op    95396 allocs/op
BenchmarkReadObject_citm_catalog/rjson-8             132	 9121031 ns/op	189.36 MB/s  4897968 B/op    76435 allocs/op
BenchmarkReadObject_citm_catalog/gjson-8              72	15723326 ns/op	109.85 MB/s  6433823 B/op    54373 allocs/op
BenchmarkReadObject_citm_catalog/jsoniter-8          100	10662080 ns/op	162.00 MB/s  5571756 B/op   118767 allocs/op

BenchmarkReadObject_github_repo/encoding_json-8    13196	   92380 ns/op	 66.41 MB/s    25687 B/op      503 allocs/op
BenchmarkReadObject_github_repo/rjson-8            30873	   37582 ns/op	163.24 MB/s    50397 B/op      322 allocs/op
BenchmarkReadObject_github_repo/gjson-8            38083	   31592 ns/op	194.20 MB/s    22676 B/op      122 allocs/op
BenchmarkReadObject_github_repo/jsoniter-8         26952	   43633 ns/op	140.61 MB/s    25462 B/op      531 allocs/op

BenchmarkReadObject_twitter/encoding_json-8          164	 7314738 ns/op	 86.33 MB/s  2152110 B/op    31267 allocs/op
BenchmarkReadObject_twitter/rjson-8                  313	 3780066 ns/op	167.06 MB/s  2970330 B/op    28473 allocs/op
BenchmarkReadObject_twitter/gjson-8                  187	 6426500 ns/op	 98.27 MB/s  2395553 B/op    12182 allocs/op
BenchmarkReadObject_twitter/jsoniter-8               226	 5091349 ns/op	124.04 MB/s  2426808 B/op    45044 allocs/op
```

### Repo Values

This tests getting three values of different types from github_repo.json.

```
BenchmarkGetRepoValues/encoding_json-8    30466     39900 ns/op     248 B/op     7 allocs/op
BenchmarkGetRepoValues/rjson-8           576297      2076 ns/op      16 B/op     1 allocs/op
BenchmarkGetRepoValues/gjson-8           436257      2339 ns/op     200 B/op     4 allocs/op
BenchmarkGetRepoValues/jsoniter-8        486663      2520 ns/op     360 B/op    25 allocs/op
BenchmarkGetRepoValues/jsonparser-8      488737      2462 ns/op      16 B/op     1 allocs/op
BenchmarkGetRepoValues/fastjson-8        235377      5089 ns/op      16 B/op     1 allocs/op
```

### Validating JSON

This benchmark is for each package's equivalent of running `encoding_json.Valid()`.

```
BenchmarkValid_canada/encoding_json-8           172	   7093858 ns/op    317.32 MB/s         7 B/op        0 allocs/op
BenchmarkValid_canada/rjson-8                   505	   2311145 ns/op    974.00 MB/s         0 B/op        0 allocs/op
BenchmarkValid_canada/gjson-8                   482	   2468825 ns/op    911.79 MB/s         0 B/op        0 allocs/op
BenchmarkValid_canada/jsoniter-8                232	   5058736 ns/op    444.98 MB/s        69 B/op        8 allocs/op
BenchmarkValid_canada/fastjson-8                351	   3387308 ns/op    664.55 MB/s         0 B/op        0 allocs/op

BenchmarkValid_citm_catalog/encoding_json-8     225	   5380901 ns/op    320.99 MB/s         5 B/op        0 allocs/op
BenchmarkValid_citm_catalog/rjson-8             727	   1673708 ns/op   1031.96 MB/s         0 B/op        0 allocs/op
BenchmarkValid_citm_catalog/gjson-8             696	   1680964 ns/op   1027.51 MB/s         0 B/op        0 allocs/op
BenchmarkValid_citm_catalog/jsoniter-8          418	   2970893 ns/op    581.38 MB/s    240578 B/op    25874 allocs/op
BenchmarkValid_citm_catalog/fastjson-8          681	   1681480 ns/op   1027.19 MB/s         0 B/op        0 allocs/op

BenchmarkValid_github_repo/encoding_json-8    55089	     21353 ns/op    287.31 MB/s         0 B/op        0 allocs/op
BenchmarkValid_github_repo/rjson-8           218859	      5576 ns/op   1100.27 MB/s         0 B/op        0 allocs/op
BenchmarkValid_github_repo/gjson-8           142514	      8264 ns/op    742.33 MB/s         0 B/op        0 allocs/op
BenchmarkValid_github_repo/jsoniter-8        101544	     11596 ns/op    529.05 MB/s      1576 B/op      122 allocs/op
BenchmarkValid_github_repo/fastjson-8        162109	      7378 ns/op    831.54 MB/s         0 B/op        0 allocs/op

BenchmarkValid_twitter/encoding_json-8          517	   2273221 ns/op    277.81 MB/s         2 B/op        0 allocs/op
BenchmarkValid_twitter/rjson-8                 1868	    622671 ns/op   1014.20 MB/s         0 B/op        0 allocs/op
BenchmarkValid_twitter/gjson-8                 1460	    803977 ns/op    785.49 MB/s         0 B/op        0 allocs/op
BenchmarkValid_twitter/jsoniter-8               709	   1656385 ns/op    381.26 MB/s    375920 B/op    15274 allocs/op
BenchmarkValid_twitter/fastjson-8              1358	    853157 ns/op    740.21 MB/s         0 B/op        0 allocs/op
```
