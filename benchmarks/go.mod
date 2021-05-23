module github.com/willabides/rjson/benchmarks

go 1.16

replace github.com/willabides/rjson => ./..

require (
	github.com/buger/jsonparser v1.1.1
	github.com/goccy/go-json v0.4.11
	github.com/json-iterator/go v1.1.10
	github.com/minio/simdjson-go v0.2.1
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.7.4
	github.com/valyala/fastjson v1.6.3
	github.com/willabides/benchdiff v0.6.2
	github.com/willabides/rjson v0.0.0
	golang.org/x/perf v0.0.0-20201207232921-bdcc6220ee90
)
