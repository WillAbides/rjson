package benchmarks

import (
	"encoding/json"
	"errors"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/willabides/rjson"
)

type readObjectRunner struct {
	name string
	fn   func(data []byte, pool *sync.Pool) (map[string]interface{}, error)
}

var readObjectRunners = []readObjectRunner{
	{
		name: "encoding_json",
		fn:   encodingJSONReadObject,
	},
	{
		name: "rjson",
		fn:   rjsonReadObject,
	},
	{
		name: "jsoniter",
		fn:   jsoniterReadObject,
	},
	{
		name: "gjson",
		fn:   gjsonReadObject,
	},
}

// ******** encoding/json ********

func encodingJSONReadObject(data []byte, _ *sync.Pool) (map[string]interface{}, error) {
	var val map[string]interface{}
	err := json.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ******** rjson ********

func rjsonReadObject(data []byte, pool *sync.Pool) (map[string]interface{}, error) {
	h, ok := pool.Get().(*rjson.ValueReader)
	if !ok {
		h = &rjson.ValueReader{}
	}
	val, _, err := h.ReadObject(data)
	pool.Put(h)
	return val, err
}

// ******** gjson ********

func gjsonReadObject(data []byte, _ *sync.Pool) (map[string]interface{}, error) {
	result := gjson.ParseBytes(data)
	mp, ok := result.Value().(map[string]interface{})
	if !ok {
		return nil, errors.New("not a map")
	}
	return mp, nil
}

// ******** jsoniter ********

func jsoniterReadObject(data []byte, pool *sync.Pool) (map[string]interface{}, error) {
	iter, ok := pool.Get().(*jsoniter.Iterator)
	if !ok {
		iter = jsoniter.NewIterator(jsoniter.ConfigFastest)
	}
	iter.ResetBytes(data)
	var val map[string]interface{}
	iter.ReadVal(&val)
	err := iter.Error
	iter.Attachment = nil
	iter.Error = nil
	pool.Put(iter)
	if err != nil {
		return nil, err
	}
	return val, nil
}
