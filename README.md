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

#### DistinctUserIDs

DistinctUserIDs returns a slice of user ids from `twitter.json`.

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
| GetRepoValues            |      39772.3 |       2156.9 | -94.58% |
| Valid_canada             |      6981750 |      2413060 | -65.44% |
| Valid_citm_catalog       |      5380290 |      1645390 | -69.42% |
| Valid_github_repo        |        19912 |       5485.7 | -72.45% |
| Valid_twitter            |      2141580 |       611203 | -71.46% |
| DistinctUserIDs          |      4227210 |       743022 | -82.42% |
| ReadObject_canada        |     52882000 |     32002000 | -39.48% |
| ReadObject_citm_catalog  |     22354300 |     11808100 | -47.18% |
| ReadObject_github_repo   |      92965.4 |      30023.4 | -67.70% |
| ReadObject_twitter       |      9228560 |      4593510 | -50.23% |
| DecodeInt64_zero         |      328.378 |       13.328 | -95.94% |
| DecodeInt64_small        |      360.122 |       14.794 | -95.89% |
| DecodeInt64_big_negative |        563.4 |       29.224 | -94.81% |

|           name           | encoding_json (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         250 |          16 | -93.60%  |
| Valid_canada             |      145800 |       47619 | -67.34%  |
| Valid_citm_catalog       |     86129.9 |       24020 | -72.11%  |
| Valid_github_repo        |           1 |           0 | -100.00% |
| Valid_twitter            |     12613.7 |      3345.5 | -73.48%  |
| DistinctUserIDs          |       26794 |     4072.67 | -84.80%  |
| ReadObject_canada        |    12070600 |     8434380 | -30.12%  |
| ReadObject_citm_catalog  |     5468880 |     5229700 | -4.37%   |
| ReadObject_github_repo   |     25704.1 |     17036.9 | -33.72%  |
| ReadObject_twitter       |     2190850 |     2717220 | +24.03%  |
| DecodeInt64_zero         |         184 |           0 | -100.00% |
| DecodeInt64_small        |         186 |           0 | -100.00% |
| DecodeInt64_big_negative |         208 |           0 | -100.00% |

### fastjson

|          name           | fastjson (ns/op) | rjson (ns/op) |  delta   |
|-------------------------|-------------:|-------------:|----------|
| GetRepoValues           |      4919.22 |       2156.9 | -56.15%  |
| ReadFloat64_zero        |      28.9222 |       14.504 | -49.85%  |
| ReadFloat64_smallInt    |       30.047 |       17.713 | -41.05%  |
| ReadFloat64_negExp      |        43.04 |      27.6944 | -35.65%  |
| ReadInt64_zero          |       28.564 |       12.406 | -56.57%  |
| ReadInt64_small         |       29.473 |       13.144 | -55.40%  |
| ReadInt64_big_negative  |       89.229 |       27.828 | -68.81%  |
| ReadString_short_ascii  |        62.68 |        39.98 | -36.22%  |
| ReadString_medium_ascii |      574.967 |       1872.8 | +225.72% |
| ReadString_medium       |        301.4 |       817.07 | +171.09% |
| ReadBool                |        19.39 |       9.0604 | -53.27%  |

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
| GetRepoValues            |      2408.33 |       2156.9 | -10.44% |
| Valid_canada             |      2531630 |      2413060 | -4.68%  |
| Valid_citm_catalog       |      1665450 |      1645390 | ~       |
| Valid_github_repo        |       7892.7 |       5485.7 | -30.50% |
| Valid_twitter            |       806403 |       611203 | -24.21% |
| DistinctUserIDs          |      1471150 |       743022 | -49.49% |
| ReadObject_canada        |     64774800 |     32002000 | -50.59% |
| ReadObject_citm_catalog  |     16946300 |     11808100 | -30.32% |
| ReadObject_github_repo   |      31775.6 |      30023.4 | -5.51%  |
| ReadObject_twitter       |      7668040 |      4593510 | -40.10% |
| ReadFloat64_zero         |      45.7111 |       14.504 | -68.27% |
| ReadFloat64_smallInt     |       55.236 |       17.713 | -67.93% |
| ReadFloat64_negExp       |       89.102 |      27.6944 | -68.92% |
| ReadInt64_zero           |       51.251 |       12.406 | -75.79% |
| ReadInt64_small          |       62.192 |       13.144 | -78.87% |
| ReadInt64_big_negative   |       156.53 |       27.828 | -82.22% |
| DecodeInt64_zero         |       52.371 |       13.328 | -74.55% |
| DecodeInt64_small        |       60.197 |       14.794 | -75.42% |
| DecodeInt64_big_negative |      155.967 |       29.224 | -81.26% |
| ReadString_short_ascii   |       106.93 |        39.98 | -62.61% |
| ReadString_medium_ascii  |      3896.33 |       1872.8 | -51.93% |
| ReadString_medium        |       1669.1 |       817.07 | -51.05% |
| ReadBool                 |       44.706 |       9.0604 | -79.73% |

|           name           | gjson (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         200 |          16 | -92.00%  |
| Valid_canada             |     49880.1 |       47619 | -4.53%   |
| Valid_citm_catalog       |     24880.3 |       24020 | -3.46%   |
| Valid_github_repo        |           0 |           0 | ~        |
| Valid_twitter            |     4417.56 |      3345.5 | -24.27%  |
| DistinctUserIDs          |     32438.1 |     4072.67 | -87.44%  |
| ReadObject_canada        |    11677900 |     8434380 | -27.78%  |
| ReadObject_citm_catalog  |     6697300 |     5229700 | -21.91%  |
| ReadObject_github_repo   |     22679.1 |     17036.9 | -24.88%  |
| ReadObject_twitter       |     2431030 |     2717220 | +11.77%  |
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
| GetRepoValues            |      2565.33 |       2156.9 | -15.92% |
| DistinctUserIDs          |      1661190 |       743022 | -55.27% |
| ReadObject_canada        |     73041000 |     32002000 | -56.19% |
| ReadObject_citm_catalog  |     13473400 |     11808100 | -12.36% |
| ReadObject_github_repo   |      45984.9 |      30023.4 | -34.71% |
| ReadObject_twitter       |      6713870 |      4593510 | -31.58% |
| ReadFloat64_zero         |       107.57 |       14.504 | -86.52% |
| ReadFloat64_smallInt     |       112.99 |       17.713 | -84.32% |
| ReadFloat64_negExp       |       149.69 |      27.6944 | -81.50% |
| ReadInt64_zero           |       14.027 |       12.406 | -11.56% |
| ReadInt64_small          |        21.35 |       13.144 | -38.44% |
| ReadInt64_big_negative   |      46.1238 |       27.828 | -39.67% |
| DecodeInt64_zero         |       99.286 |       13.328 | -86.58% |
| DecodeInt64_small        |       105.21 |       14.794 | -85.94% |
| DecodeInt64_big_negative |       134.32 |       29.224 | -78.24% |
| ReadString_short_ascii   |       50.339 |        39.98 | -20.58% |
| ReadString_medium_ascii  |       4839.3 |       1872.8 | -61.30% |
| ReadString_medium        |       2226.4 |       817.07 | -63.30% |
| ReadBool                 |      24.8511 |       9.0604 | -63.54% |

|           name           | jsoniter (B/op) | rjson (B/op) |  delta   |
|--------------------------|------------:|------------:|----------|
| GetRepoValues            |         360 |          16 | -95.56%  |
| DistinctUserIDs          |      385571 |     4072.67 | -98.94%  |
| ReadObject_canada        |    17315200 |     8434380 | -51.29%  |
| ReadObject_citm_catalog  |     5787990 |     5229700 | -9.65%   |
| ReadObject_github_repo   |     25469.3 |     17036.9 | -33.11%  |
| ReadObject_twitter       |     2452640 |     2717220 | +10.79%  |
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
| GetRepoValues           |      2447.22 |       2156.9 | -11.86% |
| DistinctUserIDs         |      1022040 |       743022 | -27.30% |
| ReadFloat64_zero        |       43.557 |       14.504 | -66.70% |
| ReadFloat64_smallInt    |      46.0313 |       17.713 | -61.52% |
| ReadFloat64_negExp      |      67.8511 |      27.6944 | -59.18% |
| ReadInt64_zero          |      25.7944 |       12.406 | -51.90% |
| ReadInt64_small         |      27.1367 |       13.144 | -51.56% |
| ReadInt64_big_negative  |        59.38 |       27.828 | -53.14% |
| ReadString_short_ascii  |      49.6714 |        39.98 | -19.51% |
| ReadString_medium_ascii |      1806.89 |       1872.8 | +3.65%  |
| ReadString_medium       |       811.09 |       817.07 | ~       |
| ReadBool                |       30.078 |       9.0604 | -69.88% |

|          name           | jsonparser (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |          16 |          16 | ~       |
| DistinctUserIDs         |      6170.9 |     4072.67 | -34.00% |
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
| GetRepoValues           |      7488.44 |       2156.9 | -71.20% |
| ReadObject_canada       |     47965900 |     32002000 | -33.28% |
| ReadObject_citm_catalog |     15102800 |     11808100 | -21.82% |
| ReadObject_github_repo  |      43846.8 |      30023.4 | -31.53% |
| ReadObject_twitter      |     14948500 |      4593510 | -69.27% |

|          name           | goccyjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |        6152 |          16 | -99.74% |
| ReadObject_canada       |    10980000 |     8434380 | -23.18% |
| ReadObject_citm_catalog |     7411250 |     5229700 | -29.44% |
| ReadObject_github_repo  |     26592.9 |     17036.9 | -35.93% |
| ReadObject_twitter      |     2811040 |     2717220 | -3.34%  |

### simdjson

|          name           | simdjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      11510.1 |       2156.9 | -81.26% |
| DistinctUserIDs         |       992518 |       743022 | -25.14% |
| ReadObject_canada       |     30483200 |     32002000 | +4.98%  |
| ReadObject_citm_catalog |     10949300 |     11808100 | +7.84%  |
| ReadObject_github_repo  |      36584.2 |      30023.4 | -17.93% |
| ReadObject_twitter      |      4399670 |      4593510 | +4.41%  |

|          name           | simdjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |         712 |          16 | -97.75% |
| DistinctUserIDs         |     21623.2 |     4072.67 | -81.17% |
| ReadObject_canada       |    12702400 |     8434380 | -33.60% |
| ReadObject_citm_catalog |     8301120 |     5229700 | -37.00% |
| ReadObject_github_repo  |     23330.7 |     17036.9 | -26.98% |
| ReadObject_twitter      |     2652500 |     2717220 | +2.44%  |
