# rjson

[![godoc](https://pkg.go.dev/badge/github.com/willabides/rjson.svg)](https://pkg.go.dev/github.com/willabides/rjson)
[![ci](https://github.com/WillAbides/rjson/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/rjson/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

rjson is a json parser that relies on [Ragel-generated](http://www.colm.net/open-source/ragel/) state machines for most
parsing. rjson's api is minimal and focussed on efficient parsing.

The center pieces of the api are HandleObjectValues and HandleArrayValues which will run a handler on each value in a
json object or array.

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

[source](benchmarks/read_object.go)

This measures the ability to create a `map[string]interface{}` from a JSON object. That isn't an operation you want
to do often where performance matters, but ironically it makes a pretty good comparative benchmark because it covers
parsing of several types.

rjson is fastest for all of these. The largest margin is on canada.json where all the packages had their slowest
performance.

```
BenchmarkReadObject_canada/encoding_json         	      68	  52140810 ns/op	  43.17 MB/s	10943944 B/op	  392541 allocs/op
BenchmarkReadObject_canada/rjson                 	      78	  40103443 ns/op	  56.13 MB/s	 7890720 B/op	  223656 allocs/op
BenchmarkReadObject_canada/jsoniter              	      55	  69744361 ns/op	  32.28 MB/s	16188593 B/op	  665994 allocs/op
BenchmarkReadObject_canada/gjson                 	      56	  58090976 ns/op	  38.75 MB/s	10551289 B/op	  281393 allocs/op

BenchmarkReadObject_citm_catalog/encoding_json   	     160	  22330150 ns/op	  77.35 MB/s	 5123197 B/op	   95396 allocs/op
BenchmarkReadObject_citm_catalog/rjson           	     315	  11426571 ns/op	 151.16 MB/s	 4856575 B/op	   76431 allocs/op
BenchmarkReadObject_citm_catalog/jsoniter        	     264	  13634860 ns/op	 126.68 MB/s	 5571853 B/op	  118769 allocs/op
BenchmarkReadObject_citm_catalog/gjson           	     194	  18326577 ns/op	  94.25 MB/s	 6433798 B/op	   54373 allocs/op

BenchmarkReadObject_github_repo/encoding_json    	   39589	     91116 ns/op	  67.33 MB/s	   25696 B/op	     503 allocs/op
BenchmarkReadObject_github_repo/rjson            	  122584	     29116 ns/op	 210.71 MB/s	   17035 B/op	     318 allocs/op
BenchmarkReadObject_github_repo/jsoniter         	   79528	     43325 ns/op	 141.60 MB/s	   25463 B/op	     531 allocs/op
BenchmarkReadObject_github_repo/gjson            	  116310	     30788 ns/op	 199.26 MB/s	   22677 B/op	     122 allocs/op

BenchmarkReadObject_twitter/encoding_json        	     386	   9293207 ns/op	  67.95 MB/s	 2152295 B/op	   31267 allocs/op
BenchmarkReadObject_twitter/rjson                	     793	   4553240 ns/op	 138.70 MB/s	 2610559 B/op	   28457 allocs/op
BenchmarkReadObject_twitter/jsoniter             	     536	   6703885 ns/op	  94.20 MB/s	 2427138 B/op	   45046 allocs/op
BenchmarkReadObject_twitter/gjson                	     476	   7464527 ns/op	  84.60 MB/s	 2395997 B/op	   12183 allocs/op
```

### Repo Values

[source](benchmarks/repo_values.go)

This tests getting three values of different types from github_repo.json.

rjson barely edges jsonparser on this one. The margin is a mere 59ns or 3%.

```
BenchmarkGetRepoValues/encoding_json         	   89881	     39048 ns/op	     240 B/op	       6 allocs/op
BenchmarkGetRepoValues/rjson                 	 1830405	      1965 ns/op	      16 B/op	       1 allocs/op
BenchmarkGetRepoValues/jsoniter              	 1368921	      2638 ns/op	     360 B/op	      25 allocs/op
BenchmarkGetRepoValues/gjson                 	 1549179	      2342 ns/op	     200 B/op	       4 allocs/op
BenchmarkGetRepoValues/jsonparser            	 1755386	      2024 ns/op	      16 B/op	       1 allocs/op
BenchmarkGetRepoValues/fastjson              	  718406	      4879 ns/op	      16 B/op	       1 allocs/op
```

### Validating JSON

[source](benchmarks/valid.go)

This benchmark is for each package's equivalent of running `encoding_json.Valid()`.

gjson is the fastest on canada.json, and rjson is fastest on the rest.

```
BenchmarkValid_canada/encoding_json         	     477	   7498668 ns/op	 300.19 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_canada/rjson                 	     946	   3818793 ns/op	 589.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_canada/jsoniter              	     799	   4392695 ns/op	 512.46 MB/s	      64 B/op	       8 allocs/op
BenchmarkValid_canada/gjson                 	    1392	   2507717 ns/op	 897.65 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_canada/fastjson              	    1087	   3312799 ns/op	 679.50 MB/s	       0 B/op	       0 allocs/op

BenchmarkValid_citm_catalog/encoding_json   	     675	   5372861 ns/op	 321.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_citm_catalog/rjson           	    2182	   1666661 ns/op	1036.33 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_citm_catalog/jsoniter        	    1226	   2830198 ns/op	 610.28 MB/s	  240544 B/op	   25874 allocs/op
BenchmarkValid_citm_catalog/gjson           	    2122	   1679373 ns/op	1028.48 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_citm_catalog/fastjson        	    2450	   1460279 ns/op	1182.79 MB/s	       0 B/op	       0 allocs/op

BenchmarkValid_github_repo/encoding_json    	  168160	     21468 ns/op	 285.78 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_github_repo/rjson            	  674563	      5274 ns/op	1163.36 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_github_repo/jsoniter         	  306826	     11497 ns/op	 533.63 MB/s	    1576 B/op	     122 allocs/op
BenchmarkValid_github_repo/gjson            	  449780	      7922 ns/op	 774.42 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_github_repo/fastjson         	  498088	      7246 ns/op	 846.71 MB/s	       0 B/op	       0 allocs/op

BenchmarkValid_twitter/encoding_json        	    1554	   2244884 ns/op	 281.31 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_twitter/rjson                	    5822	    601644 ns/op	1049.65 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_twitter/jsoniter             	    2130	   1670190 ns/op	 378.11 MB/s	  375851 B/op	   15274 allocs/op
BenchmarkValid_twitter/gjson                	    4371	    808588 ns/op	 781.01 MB/s	       0 B/op	       0 allocs/op
BenchmarkValid_twitter/fastjson             	    4368	    813736 ns/op	 776.07 MB/s	       0 B/op	       0 allocs/op
```
