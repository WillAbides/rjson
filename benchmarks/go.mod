module github.com/willabides/rjson/benchmarks

go 1.15

require (
	github.com/json-iterator/go v1.1.10
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/willabides/rjson v0.0.0
)

replace github.com/willabides/rjson => ./..
