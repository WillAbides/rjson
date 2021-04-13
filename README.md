# rjson

[![godoc](https://pkg.go.dev/badge/github.com/willabides/rjson.svg)](https://pkg.go.dev/github.com/willabides/rjson)
[![ci](https://github.com/WillAbides/rjson/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/rjson/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

rjson is a json parser that relies on [Ragel-generated](http://www.colm.net/open-source/ragel/) state machines for most
parsing. rjson's api is minimal and focussed on efficient parsing.

## Why you shouldn't use rjson most of the time

While rjson is much faster than "encoding/json", it is faster at the cost of being more difficult to use. If you use 
rjson, your code will be more verbose and less readable than if you used the standard library. In most cases, this 
will cost more in developer time than you will make up for in time saved parsing json.

To make rjson worthwhile, one of these should apply:

- You have a very high volume of json to parse.
- You need very low latency but are stuck using json serialization.

rjson is originally written for a project where we have to process thousands of json documents per second. This is 
the sort of project that should be using rjson.

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
| GetRepoValues           |      38913.3 |       2056.6 | -94.71% |
| Valid_canada            |      6484170 |      2206670 | -65.97% |
| Valid_citm_catalog      |      5185540 |      1563420 | -69.85% |
| Valid_github_repo       |      20420.8 |       5276.7 | -74.16% |
| Valid_twitter           |      2160090 |       566341 | -73.78% |
| ReadObject_canada       |     45105700 |     25741500 | -42.93% |
| ReadObject_citm_catalog |     17937000 |     11083500 | -38.21% |
| ReadObject_github_repo  |        93517 |      32133.7 | -65.64% |
| ReadObject_twitter      |      7811660 |      4160820 | -46.74% |

|          name           | encoding_json (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         250 |          16 | -93.60%  |
| Valid_canada            |      140837 |     44592.4 | -68.34%  |
| Valid_citm_catalog      |       82338 |     23178.2 | -71.85%  |
| Valid_github_repo       |           1 |           0 | -100.00% |
| Valid_twitter           |     12751.9 |      3207.3 | -74.85%  |
| ReadObject_canada       |    11695000 |     8433900 | -27.88%  |
| ReadObject_citm_catalog |     5411520 |     5114650 | -5.49%   |
| ReadObject_github_repo  |     25698.7 |     17038.1 | -33.70%  |
| ReadObject_twitter      |     2194880 |     2690170 | +22.57%  |

### fastjson

|          name           | fastjson (ns/op) | rjson (ns/op) |  delta   |
|-------------------------|-------------:|-------------:|----------|
| GetRepoValues           |      4693.56 |       2056.6 | -56.18%  |
| ReadFloat64_zero        |      27.3833 |       13.686 | -50.02%  |
| ReadFloat64_smallInt    |        27.99 |       16.733 | -40.22%  |
| ReadFloat64_negExp      |      40.5811 |      25.7811 | -36.47%  |
| ReadInt64_zero          |       27.284 |      11.0838 | -59.38%  |
| ReadInt64_small         |       27.883 |       13.504 | -51.57%  |
| ReadInt64_big_negative  |      83.9378 |        26.88 | -67.98%  |
| ReadString_short_ascii  |       58.306 |        32.86 | -43.64%  |
| ReadString_medium_ascii |       635.23 |       1746.2 | +174.89% |
| ReadString_medium       |       317.56 |       738.61 | +132.59% |
| ReadBool                |       18.936 |       8.0014 | -57.75%  |

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

|          name           | gjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      2385.78 |       2056.6 | -13.80% |
| Valid_canada            |      2382180 |      2206670 | -7.37%  |
| Valid_citm_catalog      |      1610320 |      1563420 | ~       |
| Valid_github_repo       |       7454.2 |       5276.7 | -29.21% |
| Valid_twitter           |       761576 |       566341 | -25.64% |
| ReadObject_canada       |     57734500 |     25741500 | -55.41% |
| ReadObject_citm_catalog |     17878600 |     11083500 | -38.01% |
| ReadObject_github_repo  |      36293.7 |      32133.7 | -11.46% |
| ReadObject_twitter      |      6869820 |      4160820 | -39.43% |
| ReadFloat64_zero        |       43.645 |       13.686 | -68.64% |
| ReadFloat64_smallInt    |       52.223 |       16.733 | -67.96% |
| ReadFloat64_negExp      |       85.052 |      25.7811 | -69.69% |
| ReadInt64_zero          |       47.348 |      11.0838 | -76.59% |
| ReadInt64_small         |       56.248 |       13.504 | -75.99% |
| ReadInt64_big_negative  |       147.38 |        26.88 | -81.76% |
| ReadString_short_ascii  |       91.589 |        32.86 | -64.12% |
| ReadString_medium_ascii |       3311.4 |       1746.2 | -47.27% |
| ReadString_medium       |       1509.4 |       738.61 | -51.07% |
| ReadBool                |       47.294 |       8.0014 | -83.08% |

|          name           | gjson (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         200 |          16 | -92.00%  |
| Valid_canada            |     48199.9 |     44592.4 | -7.48%   |
| Valid_citm_catalog      |     25049.3 |     23178.2 | -7.47%   |
| Valid_github_repo       |           0 |           0 | ~        |
| Valid_twitter           |     4172.25 |      3207.3 | -23.13%  |
| ReadObject_canada       |    11677900 |     8433900 | -27.78%  |
| ReadObject_citm_catalog |     6693100 |     5114650 | -23.58%  |
| ReadObject_github_repo  |     22680.3 |     17038.1 | -24.88%  |
| ReadObject_twitter      |     2429740 |     2690170 | +10.72%  |
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
| GetRepoValues           |       2447.4 |       2056.6 | -15.97% |
| ReadObject_canada       |     62529400 |     25741500 | -58.83% |
| ReadObject_citm_catalog |     12246000 |     11083500 | -9.49%  |
| ReadObject_github_repo  |      50439.3 |      32133.7 | -36.29% |
| ReadObject_twitter      |      5633380 |      4160820 | -26.14% |
| ReadFloat64_zero        |        86.41 |       13.686 | -84.16% |
| ReadFloat64_smallInt    |       98.254 |       16.733 | -82.97% |
| ReadFloat64_negExp      |        131.6 |      25.7811 | -80.41% |
| ReadInt64_zero          |       13.726 |      11.0838 | -19.25% |
| ReadInt64_small         |      20.3125 |       13.504 | -33.52% |
| ReadInt64_big_negative  |       45.837 |        26.88 | -41.36% |
| ReadString_short_ascii  |       40.996 |        32.86 | -19.85% |
| ReadString_medium_ascii |       4331.1 |       1746.2 | -59.68% |
| ReadString_medium       |       1988.7 |       738.61 | -62.86% |
| ReadBool                |      23.4722 |       8.0014 | -65.91% |

|          name           | jsoniter (B/op) | rjson (B/op) |  delta   |
|-------------------------|------------:|------------:|----------|
| GetRepoValues           |         360 |          16 | -95.56%  |
| ReadObject_canada       |    17315100 |     8433900 | -51.29%  |
| ReadObject_citm_catalog |     5764010 |     5114650 | -11.27%  |
| ReadObject_github_repo  |     25469.9 |     17038.1 | -33.10%  |
| ReadObject_twitter      |     2454050 |     2690170 | +9.62%   |
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
| GetRepoValues           |         2289 |       2056.6 | -10.15% |
| ReadFloat64_zero        |       42.007 |       13.686 | -67.42% |
| ReadFloat64_smallInt    |      44.6211 |       16.733 | -62.50% |
| ReadFloat64_negExp      |        64.07 |      25.7811 | -59.76% |
| ReadInt64_zero          |       24.623 |      11.0838 | -54.99% |
| ReadInt64_small         |       26.106 |       13.504 | -48.27% |
| ReadInt64_big_negative  |       55.663 |        26.88 | -51.71% |
| ReadString_short_ascii  |       46.104 |        32.86 | -28.73% |
| ReadString_medium_ascii |       1960.9 |       1746.2 | -10.95% |
| ReadString_medium       |       800.74 |       738.61 | -7.76%  |
| ReadBool                |       27.761 |       8.0014 | -71.18% |

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
| GetRepoValues           |       6780.2 |       2056.6 | -69.67% |
| ReadObject_canada       |     31277700 |     25741500 | -17.70% |
| ReadObject_citm_catalog |     13470600 |     11083500 | -17.72% |
| ReadObject_github_repo  |      48540.7 |      32133.7 | -33.80% |
| ReadObject_twitter      |     13315800 |      4160820 | -68.75% |

|          name           | goccyjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |        6152 |          16 | -99.74% |
| ReadObject_canada       |     7767430 |     8433900 | +8.58%  |
| ReadObject_citm_catalog |     7140020 |     5114650 | -28.37% |
| ReadObject_github_repo  |     26588.8 |     17038.1 | -35.92% |
| ReadObject_twitter      |     2798680 |     2690170 | -3.88%  |

### simdjson

|          name           | simdjson (ns/op) | rjson (ns/op) |  delta  |
|-------------------------|-------------:|-------------:|---------|
| GetRepoValues           |      10905.3 |       2056.6 | -81.14% |
| ReadObject_canada       |     33910400 |     25741500 | -24.09% |
| ReadObject_citm_catalog |     10772800 |     11083500 | +2.88%  |
| ReadObject_github_repo  |      42289.4 |      32133.7 | -24.01% |
| ReadObject_twitter      |      4343340 |      4160820 | -4.20%  |

|          name           | simdjson (B/op) | rjson (B/op) |  delta  |
|-------------------------|------------:|------------:|---------|
| GetRepoValues           |         821 |          16 | -98.05% |
| ReadObject_canada       |    16547900 |     8433900 | -49.03% |
| ReadObject_citm_catalog |     8735790 |     5114650 | -41.45% |
| ReadObject_github_repo  |     23466.8 |     17038.1 | -27.39% |
| ReadObject_twitter      |     2732290 |     2690170 | -1.54%  |
