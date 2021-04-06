# rjson

[![godoc](https://pkg.go.dev/badge/github.com/willabides/rjson.svg)](https://pkg.go.dev/github.com/willabides/rjson)
[![ci](https://github.com/WillAbides/rjson/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/rjson/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

rjson is a json parser that relies on [Ragel-generated](http://www.colm.net/open-source/ragel/) state machines for most
parsing. rjson's api is minimal and focussed on efficient parsing.

## Ragel state machines

This whole thing is built around a few Ragel-generated state machines. They are defined in .rl files, and the generated
code is in .rl.go files. If you peek at the generated code, beware that it doesn't look like anything intended to be
read by a human.

## Read functions

rjson provides Read functions for simple json values (strings, numbers, booleans and null). They each take a
`data []byte` argument and read the first json value in data. They return a `p int` value that is the offset of the
first byte after the value they read. When used in a handler, this `p` should be returned as the handler's `p`.

Read functions return an error for any json type other than the type they are meant to read. This includes null. So if
there is a value that could be either a string or null, be sure to use `NextTokenType` to check which one it is.

`ReadString` and `ReadStringBytes` take a `buf []byte` argument. `ReadStringBytes` returns the result appended to
`buf` and `ReadString` uses it as a byte buffer when decoding strings to avoid memory allocations. `buf` may be nil in
both cases.

## Handlers

rjson uses handlers to parse complex json values (objects and arrays). The handler will implement either
`HandleArrayValue(data []byte) (p int, err error)` or `HandleObjectValue(fieldname, data []byte) (p int, err error)`.
`data` is the document being parsed from the current position to the end of the document.

The handler has to return `p` as either 0, or the offset after the last byte of the value it handled. This is important
as it is how rjson knows where to resume parsing the document. When the handler returns 0, rjson uses
`SkipValue` to find the next offset, so it is best to write handlers that return the next offset and avoid an extra call
to `SkipValue`.

When your handler returns an error, `HandleArrayValues` and `HandleObjectValues` will immediately return the same error.
You can use this to stop parsing a document once you have found all values you are after.

See [HandleObjectValues's example](https://pkg.go.dev/github.com/willabides/rjson#example-HandleObjectValues)

## Standard Library Compatibility

rjson endeavors to decode json to the same values as the `encoding/json` functions. In the source code, most of rjson's
exported functions are followed by an unexported version that produces the same output using `encoding/json` instead
of `rjson`. These are used to ensure compatibility.

#### The RuneError compatibility exception

When decoding strings, `encoding/json` will replace bytes with values over 127 that aren't part of a valid utf8 rune
with `utf8.RuneError`. I don't think this is correct behavior, and I have not found any other go json parser that does
this, so rjson keeps those bytes as-is when decoding strings.

If you need those RuneErrors, you can use `StdLibCompatibleString`, `StdLibCompatibleSlice`, and `StdLibCompatibleMap`.

## Differential Fuzz Testing

rjson uses fuzz testing extensively to test compatibility with `encoding/json`. See [fuzzers.go](./fuzzers.go) for the
fuzz functions and [testdata/fuzz/corpus](./testdata/fuzz/corpus) for the increasingly unwieldy corpus.

## What rjson doesn't do

- **drop in replacement for "encoding/json"** - rjson doesn't aim to be a replacement for "encoding/json". It probably
  won't ever unmarshal json to a struct, and it definitely won't ever marshal json. It does one thing well, and that one
  thing is parsing json fast and with minimal memory allocations.

- **import "unsafe"** - Not that there is necessarily anything wrong with using unsafe where it is called for. It's just
  that you are better off avoiding it where possible, and it is possible here.

- **return unchecked strings** - Some json parsers use techniques like `strings.IndexByte(data, '"')` to skip to the end
  of a string and return the contents uninspected. This can be dangerous when dealing with json from untrusted sources.
  Those strings can contain control characters and other invalid json.

- **parse from an io.Reader** - I would like to implement this in the future, but it's off the table for now.

- **keep a global resource pool** - Some json packages have handy `Borrow` and `Return` functions that let you borrow
  resources and avoid new allocations. rjson leaves it up to the user to decide how to do resource pooling.

## Performance principles and tricks

### Handle each byte once

This is the main guiding principle of rjson performance. Do everything you can to avoid handling the same byte twice.

### fp.ParseJSONFloatPrefix

`fp.ParseJSONFloatPrefix` is a rewrite of `strconv.ParseFloat` that deals in bytes instead of strings and eliminates
everything not needed for parsing json numbers. This reduces the time to read numbers by about 60%. It's an internal
package for now, but I encourage anybody who is parsing json numbers to consider copying this code or writing something
similar.

### Predict object and array size based on previous values

`ReadObject` and `ReadArray` can create a lot of nested maps and slices. It would be nice if we knew how big they would
be ahead of time so we can `make([]interface{}, 0, size)`, but we can't know that until we reach the end. In place of
that, rjson keeps track of how big the values end up being and allocates values that are as big as the previous sibling
value of the same type.

### Always Be Benchmarking

Running benchmarks for every change keeps me from going too far down bad paths. Also, consistently benchmarking against
other parsers shows me where rjson can be improved.

While I'm coding I frequently run [benchdiff](https://github.com/WillAbides/benchdiff)
with `script/benchdiff --benchtime 10ms --bench WhateverBenchmark`
to get a quick comparison with the previous commit. It takes just a couple of seconds to run, and it saves a lot of time
down the road.

## Benchmarks

Benchmarks are in [./benchmarks](./benchmarks) and run with `script/compbench`. Each package that is benchmarked has its
own file that implements the various benchmarks.

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
better known packages. They don't all have the same features, so not all the compared packages are represented in every
benchmark. There are also some packages with incorrect `Valid()` functions. Those are excluded until they are fixed.

Every package should be able to put their best foot forward in the benchmarks. I have tried to write the best code for
each package in the benchmark. If anybody sees a way to improve the benchmarks for any package, please create an issue
to let me know.

This import statement has all the players seen in the benchmarks below:

```go
import (
encoding_json "encoding/json"

jsonparser "github.com/buger/jsonparser"
jsoniter "github.com/json-iterator/go"
gjson "github.com/tidwall/gjson"
fastjson "github.com/valyala/fastjson"
rjson "github.com/willabides/rjson"
goccyjson "github.com/goccy/go-json"
)
```

### Benchmarks

#### GetRepoValues

GetRepoValues gets three values of different types from github_repo.json and writes them to a struct. The same as
running `json.Unmarshal` with this struct:

```go
type repoData struct {
Archived bool   `json:"archived"`
Forks    int64  `json:"forks"`
FullName string `json:"full_name"`
}
```

#### Valid

Valid runs checks whether json data is valid. It's the equivalent of `json.Valid(data)`

#### ReadObject

ReadObject decodes json data to a `map[string]interface{}`.

#### ReadFloat64

ReadFloat64 parses a json number and returns its value as a float64.

#### ReadInt64

ReadInt64 parses a json number and returns its value as an int64.

#### ReadString

ReadString decodes a json string and returns its value as a string.

#### ReadBool

ReadBool decodes a json `true` or `false` and returns the value.

### Results

These are the results from running the benchmarks 10x and using benchstat to compare each package to rjson. You can run
them yourself by running `script/compbench`.

### encoding_json

|          name           | encoding_json (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      36514.2 |       1942.2 | -94.68% |
| Valid_canada            |      6665030 |      2268030 | -65.97% |
| Valid_citm_catalog      |      4997300 |      1575810 | -68.47% |
| Valid_github_repo       |        18758 |       5261.4 | -71.95% |
| Valid_twitter           |      2029030 |       581240 | -71.35% |
| ReadObject_canada       |     46000600 |     26239100 | -42.96% |
| ReadObject_citm_catalog |     18930100 |     10616400 | -43.92% |
| ReadObject_github_repo  |      94230.5 |      31263.4 | -66.82% |
| ReadObject_twitter      |      7535610 |      4069240 | -46.00% |

|          name           | encoding_json (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         250 |          16 | -93.60%  |
| Valid_canada            |      130822 |     43184.6 | -66.99%  |
| Valid_citm_catalog      |     82530.9 |     23599.7 | -71.41%  |
| Valid_github_repo       |           1 |           0 | -100.00% |
| Valid_twitter           |     11724.3 |      3183.4 | -72.85%  |
| ReadObject_canada       |    11695000 |     8390470 | -28.26%  |
| ReadObject_citm_catalog |     5382490 |     5114570 | -4.98%   |
| ReadObject_github_repo  |     25699.6 |     17038.1 | -33.70%  |
| ReadObject_twitter      |     2190050 |     2689880 | +22.82%  |

### fastjson

|          name           | fastjson (ns/op) | rjson (ns/op) |  delta   |
|-------------------------|-------------:|-------------:|----------|
| GetRepoValues           |       4628.1 |       1942.2 | -58.03%  |
| ReadFloat64_zero        |       27.563 |        13.72 | -50.22%  |
| ReadFloat64_smallInt    |       28.774 |      16.9056 | -41.25%  |
| ReadFloat64_negExp      |       41.029 |       25.936 | -36.79%  |
| ReadInt64_zero          |       27.079 |       12.656 | -53.26%  |
| ReadInt64_small         |       28.202 |       12.886 | -54.31%  |
| ReadInt64_big_negative  |      84.9278 |       26.869 | -68.36%  |
| ReadString_short_ascii  |       58.301 |        33.08 | -43.26%  |
| ReadString_medium_ascii |       622.46 |      1668.25 | +168.01% |
| ReadString_medium       |       304.13 |      726.778 | +138.97% |
| ReadBool                |         18.2 |       7.6498 | -57.97%  |

|          name           | fastjson (B/op) | rjson (B/op) | delta  |
|-------------------------|------------:|------------:|--------|
| GetRepoValues           |        17.3 |          16 | -7.51% |
| ReadFloat64_zero        |           0 |           0 | ~      |
| ReadFloat64_smallInt    |           0 |           0 | ~      |
| ReadFloat64_negExp      |           0 |           0 | ~      |
| ReadInt64_zero          |           0 |           0 | ~      |
| ReadInt64_small         |           0 |           0 | ~      |
| ReadInt64_big_negative  |           0 |           0 | ~      |
| ReadString_short_ascii  |           5 |           5 | ~      |
| ReadString_medium_ascii |         896 |         896 | ~      |
| ReadString_medium       |         384 |         384 | ~      |
| ReadBool                |           0 |           0 | ~      |

### gjson

|          name           | gjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |       2269.2 |       1942.2 | -14.41% |
| Valid_canada            |      2374040 |      2268030 | -4.47%  |
| Valid_citm_catalog      |      1619840 |      1575810 | ~       |
| Valid_github_repo       |      7589.11 |       5261.4 | -30.67% |
| Valid_twitter           |       772867 |       581240 | -24.79% |
| ReadObject_canada       |     56599100 |     26239100 | -53.64% |
| ReadObject_citm_catalog |     16240800 |     10616400 | -34.63% |
| ReadObject_github_repo  |        35378 |      31263.4 | -11.63% |
| ReadObject_twitter      |      6727150 |      4069240 | -39.51% |
| ReadFloat64_zero        |      42.3911 |        13.72 | -67.63% |
| ReadFloat64_smallInt    |       51.043 |      16.9056 | -66.88% |
| ReadFloat64_negExp      |      81.6922 |       25.936 | -68.25% |
| ReadInt64_zero          |       47.748 |       12.656 | -73.49% |
| ReadInt64_small         |       56.251 |       12.886 | -77.09% |
| ReadInt64_big_negative  |      141.188 |       26.869 | -80.97% |
| ReadString_short_ascii  |       88.377 |        33.08 | -62.57% |
| ReadString_medium_ascii |       3668.8 |      1668.25 | -54.53% |
| ReadString_medium       |       1512.5 |      726.778 | -51.95% |
| ReadBool                |       43.245 |       7.6498 | -82.31% |

|          name           | gjson (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         200 |          16 | -92.00%  |
| Valid_canada            |     46387.3 |     43184.6 | -6.90%   |
| Valid_citm_catalog      |     25590.6 |     23599.7 | -7.78%   |
| Valid_github_repo       |           0 |           0 | ~        |
| Valid_twitter           |      4153.9 |      3183.4 | -23.36%  |
| ReadObject_canada       |    11677900 |     8390470 | -28.15%  |
| ReadObject_citm_catalog |     6662490 |     5114570 | -23.23%  |
| ReadObject_github_repo  |     22680.2 |     17038.1 | -24.88%  |
| ReadObject_twitter      |     2427630 |     2689880 | +10.80%  |
| ReadFloat64_zero        |           0 |           0 | ~        |
| ReadFloat64_smallInt    |           2 |           0 | -100.00% |
| ReadFloat64_negExp      |          16 |           0 | -100.00% |
| ReadInt64_zero          |           0 |           0 | ~        |
| ReadInt64_small         |           2 |           0 | -100.00% |
| ReadInt64_big_negative  |          24 |           0 | -100.00% |
| ReadString_short_ascii  |          24 |           5 | -79.17%  |
| ReadString_medium_ascii |        2960 |         896 | -69.73%  |
| ReadString_medium       |        1168 |         384 | -67.12%  |
| ReadBool                |           4 |           0 | -100.00% |

### jsoniter

|          name           | jsoniter (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |       2380.6 |       1942.2 | -18.42% |
| ReadObject_canada       |     59639500 |     26239100 | -56.00% |
| ReadObject_citm_catalog |     12060400 |     10616400 | -11.97% |
| ReadObject_github_repo  |      49371.6 |      31263.4 | -36.68% |
| ReadObject_twitter      |      5506370 |      4069240 | -26.10% |
| ReadFloat64_zero        |       87.342 |        13.72 | -84.29% |
| ReadFloat64_smallInt    |       93.267 |      16.9056 | -81.87% |
| ReadFloat64_negExp      |       127.98 |       25.936 | -79.73% |
| ReadInt64_zero          |       13.653 |       12.656 | ~       |
| ReadInt64_small         |       20.706 |       12.886 | -37.77% |
| ReadInt64_big_negative  |       45.322 |       26.869 | -40.72% |
| ReadString_short_ascii  |       40.132 |        33.08 | -17.57% |
| ReadString_medium_ascii |       4255.4 |      1668.25 | -60.80% |
| ReadString_medium       |       1984.6 |      726.778 | -63.38% |
| ReadBool                |       23.386 |       7.6498 | -67.29% |

|          name           | jsoniter (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         360 |          16 | -95.56%  |
| ReadObject_canada       |    17315100 |     8390470 | -51.54%  |
| ReadObject_citm_catalog |     5729460 |     5114570 | -10.73%  |
| ReadObject_github_repo  |     25461.7 |     17038.1 | -33.08%  |
| ReadObject_twitter      |     2452320 |     2689880 | +9.69%   |
| ReadFloat64_zero        |          16 |           0 | -100.00% |
| ReadFloat64_smallInt    |          16 |           0 | -100.00% |
| ReadFloat64_negExp      |          16 |           0 | -100.00% |
| ReadInt64_zero          |           0 |           0 | ~        |
| ReadInt64_small         |           0 |           0 | ~        |
| ReadInt64_big_negative  |           0 |           0 | ~        |
| ReadString_short_ascii  |           5 |           5 | ~        |
| ReadString_medium_ascii |        2936 |         896 | -69.48%  |
| ReadString_medium       |        1400 |         384 | -72.57%  |
| ReadBool                |           0 |           0 | ~        |

### jsonparser

|          name           | jsonparser (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |       1905.4 |       1942.2 | ~       |
| ReadFloat64_zero        |       41.418 |        13.72 | -66.87% |
| ReadFloat64_smallInt    |      44.6256 |      16.9056 | -62.12% |
| ReadFloat64_negExp      |      64.5725 |       25.936 | -59.83% |
| ReadInt64_zero          |       24.306 |       12.656 | -47.93% |
| ReadInt64_small         |      25.6289 |       12.886 | -49.72% |
| ReadInt64_big_negative  |      55.6589 |       26.869 | -51.73% |
| ReadString_short_ascii  |      43.8478 |        33.08 | -24.56% |
| ReadString_medium_ascii |      1484.67 |      1668.25 | +12.37% |
| ReadString_medium       |      626.033 |      726.778 | +16.09% |
| ReadBool                |       27.609 |       7.6498 | -72.29% |

|          name           | jsonparser (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |          16 |          16 | ~       |
| ReadFloat64_zero        |           0 |           0 | ~       |
| ReadFloat64_smallInt    |           0 |           0 | ~       |
| ReadFloat64_negExp      |           0 |           0 | ~       |
| ReadInt64_zero          |           0 |           0 | ~       |
| ReadInt64_small         |           0 |           0 | ~       |
| ReadInt64_big_negative  |           0 |           0 | ~       |
| ReadString_short_ascii  |           5 |           5 | ~       |
| ReadString_medium_ascii |        1920 |         896 | -53.33% |
| ReadString_medium       |         768 |         384 | -50.00% |
| ReadBool                |           0 |           0 | ~       |

### goccyjson

|          name           | goccyjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      35891.5 |       1942.2 | -94.59% |
| ReadObject_canada       |     29399600 |     26239100 | -10.75% |
| ReadObject_citm_catalog |     12563200 |     10616400 | -15.50% |
| ReadObject_github_repo  |      48363.8 |      31263.4 | -35.36% |
| ReadObject_twitter      |     13011500 |      4069240 | -68.73% |

|          name           | goccyjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |         250 |          16 | -93.60% |
| ReadObject_canada       |     7803840 |     8390470 | +7.52%  |
| ReadObject_citm_catalog |     7128370 |     5114570 | -28.25% |
| ReadObject_github_repo  |     26595.8 |     17038.1 | -35.94% |
| ReadObject_twitter      |     2796740 |     2689880 | -3.82%  |
