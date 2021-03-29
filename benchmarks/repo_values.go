package benchmarks

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
	"github.com/willabides/rjson"
)

type repoValues struct {
	Archived bool   `json:"archived"`
	Forks    int64  `json:"forks"`
	FullName string `json:"full_name"`
}

type getRepoValuesRunner struct {
	name string
	fn   func(data []byte, pool *sync.Pool, result *repoValues) error
}

var getRepoValuesRunners = []getRepoValuesRunner{
	{
		name: "encoding_json",
		fn:   encodingJSONGetRepoValues,
	},
	{
		name: "rjson",
		fn:   rjsonGetRepoValues,
	},
	{
		name: "jsoniter",
		fn:   jsoniterGetRepoValues,
	},
	{
		name: "gjson",
		fn:   gjsonGetRepoValues,
	},
	{
		name: "jsonparser",
		fn:   jsonparserGetRepoValues,
	},
	{
		name: "fastjson",
		fn:   fastjsonGetRepoValues,
	},
}

// ******** encoding/json ********

func encodingJSONGetRepoValues(data []byte, _ *sync.Pool, result *repoValues) error {
	return json.Unmarshal(data, result)
}

// ******** rjson ********

func rjsonGetRepoValues(data []byte, pool *sync.Pool, result *repoValues) error {
	h, ok := pool.Get().(*rjsonGetValuesFromRepoHandler)
	if !ok {
		h = &rjsonGetValuesFromRepoHandler{
			doneErr: fmt.Errorf("done"),
		}
	}
	h.seenFullName, h.seenForks, h.seenArchived = false, false, false
	h.res = result
	_, err := h.buffer.HandleObjectValues(data, h)
	pool.Put(h)
	if err == h.doneErr {
		err = nil
	}
	return err
}

type rjsonGetValuesFromRepoHandler struct {
	res          *repoValues
	buffer       rjson.Buffer
	seenArchived bool
	seenForks    bool
	seenFullName bool
	doneErr      error
	stringBuf    []byte
}

func (h *rjsonGetValuesFromRepoHandler) HandleObjectValue(fieldname, data []byte) (p int, err error) {
	var tknType rjson.TokenType
	tknType, _, err = rjson.NextTokenType(data)
	isNull := tknType == rjson.NullType
	if err != nil {
		return 0, err
	}
	switch string(fieldname) {
	case `archived`:
		h.seenArchived = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.res.Archived, p, err = rjson.ReadBool(data)
	case `forks`:
		h.seenForks = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.res.Forks, p, err = rjson.ReadInt64(data)
	case `full_name`:
		h.seenFullName = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.stringBuf, p, err = rjson.ReadStringBytes(data, h.stringBuf[:0])
		h.res.FullName = string(h.stringBuf)
	default:
		p, err = h.buffer.SkipValue(data)
	}
	if err == nil && h.seenFullName && h.seenForks && h.seenArchived {
		return p, h.doneErr
	}
	return p, err
}

// ******** gjson ********

func gjsonGetRepoValues(data []byte, _ *sync.Pool, result *repoValues) error {
	results := gjson.GetManyBytes(data, `archived`, `full_name`, `forks`)

	switch results[0].Type {
	case gjson.True, gjson.False:
		result.Archived = results[0].Bool()
	case gjson.Null:
	default:
		return errors.New("unexpected type")
	}

	switch results[1].Type {
	case gjson.String:
		result.FullName = results[1].Str
	case gjson.Null:
	default:
		return errors.New("unexpected type")
	}

	switch results[2].Type {
	case gjson.Number:
		result.Forks = results[2].Int()
	case gjson.Null:
	default:
		return errors.New("unexpected type")
	}
	return nil
}

// ******** fastjson ********

func fastjsonGetRepoValues(data []byte, pool *sync.Pool, result *repoValues) error {
	parser, ok := pool.Get().(*fastjson.Parser)
	if !ok {
		parser = &fastjson.Parser{}
	}
	defer pool.Put(parser)
	val, err := parser.ParseBytes(data)
	if err != nil {
		return err
	}
	objectVal := val.Get("archived")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			result.Archived, err = objectVal.Bool()
			if err != nil {
				return err
			}
		}
	}

	objectVal = val.Get("full_name")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			sb, sbErr := objectVal.StringBytes()
			if sbErr != nil {
				return sbErr
			}
			result.FullName = string(sb)
		}
	}

	objectVal = val.Get("forks")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			result.Forks, err = objectVal.Int64()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ******** jsoniter ********

type jsoniterGetValuesFromRepoHelper struct {
	res          *repoValues
	seenArchived bool
	seenForks    bool
	seenFullName bool
}

func (h *jsoniterGetValuesFromRepoHelper) callback(it *jsoniter.Iterator, field string) bool {
	tp := it.WhatIsNext()
	switch field {
	case `archived`:
		if tp == jsoniter.NilValue {
			it.ReadNil()
			break
		}
		h.res.Archived = it.ReadBool()
		h.seenArchived = true
	case `forks`:
		if tp == jsoniter.NilValue {
			it.ReadNil()
			break
		}
		h.res.Forks = it.ReadInt64()
		h.seenForks = true
	case `full_name`:
		if tp == jsoniter.NilValue {
			it.ReadNil()
			break
		}
		h.res.FullName = it.ReadString()
		h.seenFullName = true
	default:
		it.Skip()
	}
	if h.seenForks && h.seenFullName && h.seenArchived {
		return false
	}
	return true
}

func jsoniterGetRepoValues(data []byte, pool *sync.Pool, result *repoValues) error {
	iter, ok := pool.Get().(*jsoniter.Iterator)
	if !ok {
		iter = jsoniter.NewIterator(jsoniter.ConfigFastest)
	}
	iter.ResetBytes(data)
	handler := jsoniterGetValuesFromRepoHelper{
		res: result,
	}
	iter.ReadObjectCB(handler.callback)
	if iter.Error != nil {
		return iter.Error
	}
	pool.Put(iter)
	return nil
}

// ******** jsonparser ********

type jsonparserGetValuesFromRepoHandler struct {
	res          *repoValues
	doneErr      error
	seenArchived bool
	seenForks    bool
	seenFullName bool
}

func (h *jsonparserGetValuesFromRepoHandler) callback(key, value []byte, dataType jsonparser.ValueType, _ int) error {
	var err error
	switch string(key) {
	case "archived":
		h.seenArchived = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.Archived, err = jsonparser.ParseBoolean(value)
	case "forks":
		h.seenForks = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.Forks, err = jsonparser.ParseInt(value)
	case "full_name":
		h.seenFullName = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.FullName, err = jsonparser.ParseString(value)
	}
	if err == nil && h.seenArchived && h.seenForks && h.seenFullName {
		return h.doneErr
	}
	return err
}

func jsonparserGetRepoValues(data []byte, pool *sync.Pool, result *repoValues) error {
	h, ok := pool.Get().(*jsonparserGetValuesFromRepoHandler)
	if !ok {
		h = &jsonparserGetValuesFromRepoHandler{
			doneErr: errors.New("done"),
		}
	}
	h.seenFullName, h.seenForks, h.seenArchived = false, false, false
	h.res = result
	err := jsonparser.ObjectEach(data, h.callback)
	if err == h.doneErr {
		err = nil
	}
	pool.Put(h)
	return err
}
