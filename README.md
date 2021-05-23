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
with `utf8.RuneError`. ~~I don't think this is correct behavior~~ This is probably correct behavior for the standard 
library, but rjson keeps those bytes as-is when decoding strings.

If you need standard library compatibility here, you can use `StdLibCompatibleString`, `StdLibCompatibleSlice`, and 
`StdLibCompatibleMap`.

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

## Why you should still use the standard library most of the time

While rjson is much faster than "encoding/json", it is faster at the cost of being more difficult to use. If you use
rjson, your code will be more verbose and less readable than if you used the standard library. In most cases, this
will cost more in developer time than you will make up for in time saved parsing json.

To make rjson worthwhile, one of these should apply:

- You have a very high volume of json to parse.
- You need very low latency but are stuck using json serialization.

rjson is originally written for a project where we have to process thousands of json documents per second. This is
the sort of project that should be using rjson.

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

|           name           | encoding_json (ns/op) | rjson (ns/op) |  delta  |
|--------------------------|-------------:|-------------:|---------|
| GetRepoValues            |      39600.6 |      2116.56 | -94.66% |
| Valid_canada             |      6914530 |      2315600 | -66.51% |
| Valid_citm_catalog       |      5323730 |      1642440 | -69.15% |
| Valid_github_repo        |      20250.3 |       5650.5 | -72.10% |
| Valid_twitter            |      2130420 |       611216 | -71.31% |
| DistinctUserIDs          |      4262360 |       753133 | -82.33% |
| ReadObject_canada        |     49794600 |     26946800 | -45.88% |
| ReadObject_citm_catalog  |     19002400 |     11645500 | -38.72% |
| ReadObject_github_repo   |      96357.3 |      33139.8 | -65.61% |
| ReadObject_twitter       |      7914700 |      4379160 | -44.67% |
| DecodeInt64_zero         |       282.14 |       13.195 | -95.32% |
| DecodeInt64_small        |       313.32 |       14.597 | -95.34% |
| DecodeInt64_big_negative |       483.95 |      29.2789 | -93.95% |

|           name           | encoding_json (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         250 |          16 | -93.60%  |
| Valid_canada             |      143654 |       45994 | -67.98%  |
| Valid_citm_catalog       |     86129.9 |       25665 | -70.20%  |
| Valid_github_repo        |           1 |           0 | -100.00% |
| Valid_twitter            |     12564.5 |     3354.89 | -73.30%  |
| DistinctUserIDs          |       26794 |      4134.9 | -84.57%  |
| ReadObject_canada        |    11845200 |     8433930 | -28.80%  |
| ReadObject_citm_catalog  |     5411290 |     5136060 | -5.09%   |
| ReadObject_github_repo   |     25701.6 |     17039.4 | -33.70%  |
| ReadObject_twitter       |     2193330 |     2697120 | +22.97%  |
| DecodeInt64_zero         |         184 |           0 | -100.00% |
| DecodeInt64_small        |         186 |           0 | -100.00% |
| DecodeInt64_big_negative |         208 |           0 | -100.00% |

### fastjson

|          name           | fastjson (ns/op) | rjson (ns/op) |  delta   |
|-------------------------|-------------:|-------------:|----------|
| GetRepoValues           |      4950.12 |      2116.56 | -57.24%  |
| ReadFloat64_zero        |       29.411 |      14.1838 | -51.77%  |
| ReadFloat64_smallInt    |       30.413 |      17.4567 | -42.60%  |
| ReadFloat64_negExp      |       43.728 |      27.2667 | -37.64%  |
| ReadInt64_zero          |       28.463 |       12.009 | -57.81%  |
| ReadInt64_small         |       29.327 |       13.496 | -53.98%  |
| ReadInt64_big_negative  |       88.903 |        29.12 | -67.25%  |
| ReadString_short_ascii  |        60.77 |       33.613 | -44.69%  |
| ReadString_medium_ascii |       659.16 |       1802.7 | +173.48% |
| ReadString_medium       |       324.88 |       789.83 | +143.11% |
| ReadBool                |      19.3422 |      8.09111 | -58.17%  |

|          name           | fastjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |          18 |          16 | -11.11% |
| ReadFloat64_zero        |           0 |           0 | ~       |
| ReadFloat64_smallInt    |           0 |           0 | ~       |
| ReadFloat64_negExp      |           0 |           0 | ~       |
| ReadInt64_zero          |           0 |           0 | ~       |
| ReadInt64_small         |           0 |           0 | ~       |
| ReadInt64_big_negative  |           0 |           0 | ~       |
| ReadString_short_ascii  |           5 |           5 | ~       |
| ReadString_medium_ascii |         896 |         896 | ~       |
| ReadString_medium       |         384 |         384 | ~       |
| ReadBool                |           0 |           0 | ~       |

### gjson

|           name           | gjson (ns/op) | rjson (ns/op) |  delta  |
|--------------------------|-------------:|-------------:|---------|
| GetRepoValues            |         2461 |      2116.56 | -14.00% |
| Valid_canada             |      2549620 |      2315600 | -9.18%  |
| Valid_citm_catalog       |      1690280 |      1642440 | -2.83%  |
| Valid_github_repo        |      7949.11 |       5650.5 | -28.92% |
| Valid_twitter            |       825232 |       611216 | -25.93% |
| DistinctUserIDs          |      1477430 |       753133 | -49.02% |
| ReadObject_canada        |     59939000 |     26946800 | -55.04% |
| ReadObject_citm_catalog  |     17424900 |     11645500 | -33.17% |
| ReadObject_github_repo   |      37761.7 |      33139.8 | -12.24% |
| ReadObject_twitter       |      7186710 |      4379160 | -39.07% |
| ReadFloat64_zero         |       45.976 |      14.1838 | -69.15% |
| ReadFloat64_smallInt     |       53.784 |      17.4567 | -67.54% |
| ReadFloat64_negExp       |      83.5444 |      27.2667 | -67.36% |
| ReadInt64_zero           |       51.312 |       12.009 | -76.60% |
| ReadInt64_small          |      59.3838 |       13.496 | -77.27% |
| ReadInt64_big_negative   |       150.39 |        29.12 | -80.64% |
| DecodeInt64_zero         |       51.401 |       13.195 | -74.33% |
| DecodeInt64_small        |        58.84 |       14.597 | -75.19% |
| DecodeInt64_big_negative |       152.05 |      29.2789 | -80.74% |
| ReadString_short_ascii   |       96.432 |       33.613 | -65.14% |
| ReadString_medium_ascii  |       3891.7 |       1802.7 | -53.68% |
| ReadString_medium        |      1586.17 |       789.83 | -50.21% |
| ReadBool                 |       46.949 |      8.09111 | -82.77% |

|           name           | gjson (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         200 |          16 | -92.00%  |
| Valid_canada             |       50073 |       45994 | -8.15%   |
| Valid_citm_catalog       |     26336.1 |       25665 | ~        |
| Valid_github_repo        |           0 |           0 | ~        |
| Valid_twitter            |        4420 |     3354.89 | -24.10%  |
| DistinctUserIDs          |     32273.6 |      4134.9 | -87.19%  |
| ReadObject_canada        |    11677900 |     8433930 | -27.78%  |
| ReadObject_citm_catalog  |     6680960 |     5136060 | -23.12%  |
| ReadObject_github_repo   |     22679.2 |     17039.4 | -24.87%  |
| ReadObject_twitter       |     2429560 |     2697120 | +11.01%  |
| ReadFloat64_zero         |           0 |           0 | ~        |
| ReadFloat64_smallInt     |           2 |           0 | -100.00% |
| ReadFloat64_negExp       |          16 |           0 | -100.00% |
| ReadInt64_zero           |           0 |           0 | ~        |
| ReadInt64_small          |           2 |           0 | -100.00% |
| ReadInt64_big_negative   |          24 |           0 | -100.00% |
| DecodeInt64_zero         |           0 |           0 | ~        |
| DecodeInt64_small        |           2 |           0 | -100.00% |
| DecodeInt64_big_negative |          24 |           0 | -100.00% |
| ReadString_short_ascii   |          24 |           5 | -79.17%  |
| ReadString_medium_ascii  |        2960 |         896 | -69.73%  |
| ReadString_medium        |        1168 |         384 | -67.12%  |
| ReadBool                 |           4 |           0 | -100.00% |

### jsoniter

|           name           | jsoniter (ns/op) | rjson (ns/op) |  delta  |
|--------------------------|-------------:|-------------:|---------|
| GetRepoValues            |       2603.2 |      2116.56 | -18.69% |
| DistinctUserIDs          |      1650610 |       753133 | -54.37% |
| ReadObject_canada        |     66331100 |     26946800 | -59.38% |
| ReadObject_citm_catalog  |     12828400 |     11645500 | -9.22%  |
| ReadObject_github_repo   |      52348.7 |      33139.8 | -36.69% |
| ReadObject_twitter       |      5978090 |      4379160 | -26.75% |
| ReadFloat64_zero         |      87.2467 |      14.1838 | -83.74% |
| ReadFloat64_smallInt     |       104.14 |      17.4567 | -83.24% |
| ReadFloat64_negExp       |      133.956 |      27.2667 | -79.64% |
| ReadInt64_zero           |       14.199 |       12.009 | -15.42% |
| ReadInt64_small          |      21.3222 |       13.496 | -36.70% |
| ReadInt64_big_negative   |       46.717 |        29.12 | -37.67% |
| DecodeInt64_zero         |       95.202 |       13.195 | -86.14% |
| DecodeInt64_small        |      102.132 |       14.597 | -85.71% |
| DecodeInt64_big_negative |       131.64 |      29.2789 | -77.76% |
| ReadString_short_ascii   |       40.793 |       33.613 | -17.60% |
| ReadString_medium_ascii  |       4592.7 |       1802.7 | -60.75% |
| ReadString_medium        |      2057.75 |       789.83 | -61.62% |
| ReadBool                 |       25.318 |      8.09111 | -68.04% |

|           name           | jsoniter (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         360 |          16 | -95.56%  |
| DistinctUserIDs          |      385439 |      4134.9 | -98.93%  |
| ReadObject_canada        |    17315200 |     8433930 | -51.29%  |
| ReadObject_citm_catalog  |     5756820 |     5136060 | -10.78%  |
| ReadObject_github_repo   |     25461.9 |     17039.4 | -33.08%  |
| ReadObject_twitter       |     2455040 |     2697120 | +9.86%   |
| ReadFloat64_zero         |          16 |           0 | -100.00% |
| ReadFloat64_smallInt     |          16 |           0 | -100.00% |
| ReadFloat64_negExp       |          16 |           0 | -100.00% |
| ReadInt64_zero           |           0 |           0 | ~        |
| ReadInt64_small          |           0 |           0 | ~        |
| ReadInt64_big_negative   |           0 |           0 | ~        |
| DecodeInt64_zero         |           0 |           0 | ~        |
| DecodeInt64_small        |           0 |           0 | ~        |
| DecodeInt64_big_negative |           0 |           0 | ~        |
| ReadString_short_ascii   |           5 |           5 | ~        |
| ReadString_medium_ascii  |        2936 |         896 | -69.48%  |
| ReadString_medium        |        1400 |         384 | -72.57%  |
| ReadBool                 |           0 |           0 | ~        |

### jsonparser

|          name           | jsonparser (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |       2020.7 |      2116.56 | +4.74%  |
| DistinctUserIDs         |       922930 |       753133 | -18.40% |
| ReadFloat64_zero        |       43.741 |      14.1838 | -67.57% |
| ReadFloat64_smallInt    |      45.9212 |      17.4567 | -61.99% |
| ReadFloat64_negExp      |      67.4825 |      27.2667 | -59.59% |
| ReadInt64_zero          |      25.7767 |       12.009 | -53.41% |
| ReadInt64_small         |       27.575 |       13.496 | -51.06% |
| ReadInt64_big_negative  |       59.112 |        29.12 | -50.74% |
| ReadString_short_ascii  |      48.4433 |       33.613 | -30.61% |
| ReadString_medium_ascii |      1595.62 |       1802.7 | +12.98% |
| ReadString_medium       |      662.567 |       789.83 | +19.21% |
| ReadBool                |       29.986 |      8.09111 | -73.02% |

|          name           | jsonparser (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |          16 |          16 | ~       |
| DistinctUserIDs         |      4945.5 |      4134.9 | -16.39% |
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
| GetRepoValues           |      7109.25 |      2116.56 | -70.23% |
| ReadObject_canada       |     33791900 |     26946800 | -20.26% |
| ReadObject_citm_catalog |     14062900 |     11645500 | -17.19% |
| ReadObject_github_repo  |      48828.4 |      33139.8 | -32.13% |
| ReadObject_twitter      |     13765400 |      4379160 | -68.19% |

|          name           | goccyjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |        6152 |          16 | -99.74% |
| ReadObject_canada       |     7942610 |     8433930 | +6.19%  |
| ReadObject_citm_catalog |     7146590 |     5136060 | -28.13% |
| ReadObject_github_repo  |     26591.1 |     17039.4 | -35.92% |
| ReadObject_twitter      |     2801030 |     2697120 | -3.71%  |

### simdjson

|          name           | simdjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      11782.9 |      2116.56 | -82.04% |
| DistinctUserIDs         |      1027120 |       753133 | -26.68% |
| ReadObject_canada       |     34362000 |     26946800 | -21.58% |
| ReadObject_citm_catalog |     11142600 |     11645500 | +4.51%  |
| ReadObject_github_repo  |      41890.3 |      33139.8 | -20.89% |
| ReadObject_twitter      |      4411260 |      4379160 | ~       |

|          name           | simdjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |         808 |          16 | -98.02% |
| DistinctUserIDs         |       33368 |      4134.9 | -87.61% |
| ReadObject_canada       |    15440500 |     8433930 | -45.38% |
| ReadObject_citm_catalog |     8526420 |     5136060 | -39.76% |
| ReadObject_github_repo  |     23425.4 |     17039.4 | -27.26% |
| ReadObject_twitter      |     2664980 |     2697120 | +1.21%  |
