package benchmarks

import (
	"encoding/json"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
	"github.com/willabides/rjson"
)

type validRunner struct {
	name string
	fn   func(data []byte, pool *sync.Pool) bool
}

var validRunners = []validRunner{
	{
		name: "encoding_json",
		fn:   encodingJSONValid,
	},
	{
		name: "rjson",
		fn:   rjsonValid,
	},
	{
		name: "jsoniter",
		fn:   jsoniterValid,
	},
	{
		name: "gjson",
		fn:   gjsonValid,
	},
	{
		name: "fastjson",
		fn:   fastjsonValid,
	},
}

func encodingJSONValid(data []byte, _ *sync.Pool) bool {
	return json.Valid(data)
}

func rjsonValid(data []byte, pool *sync.Pool) bool {
	buffer, ok := pool.Get().(*rjson.Buffer)
	if !ok {
		buffer = &rjson.Buffer{}
	}
	result := buffer.Valid(data)
	pool.Put(buffer)
	return result
}

func jsoniterValid(data []byte, _ *sync.Pool) bool {
	return jsoniter.Valid(data)
}

func gjsonValid(data []byte, _ *sync.Pool) bool {
	return gjson.ValidBytes(data)
}

func fastjsonValid(data []byte, _ *sync.Pool) bool {
	err := fastjson.ValidateBytes(data)
	return err == nil
}
