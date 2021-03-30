package benchmarks

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

type jsoniterBencher struct {
	iter *jsoniter.Iterator
}

func (x *jsoniterBencher) name() string {
	return "jsoniter"
}

func (x *jsoniterBencher) init() {
	x.iter = jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary)
}

func (x *jsoniterBencher) resetIter(data []byte) {
	x.iter.Attachment = nil
	x.iter.Error = nil
	x.iter.ResetBytes(data)
}

func (x *jsoniterBencher) readFloat64(data []byte) (val float64, err error) {
	x.resetIter(data)
	val = x.iter.ReadFloat64()
	err = x.iter.Error
	if err == io.EOF {
		err = nil
	}
	return val, err
}

func (x *jsoniterBencher) readInt64(data []byte) (val int64, err error) {
	x.resetIter(data)
	val = x.iter.ReadInt64()
	err = x.iter.Error
	if err == io.EOF {
		err = nil
	}
	return val, err
}

func (x *jsoniterBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	x.resetIter(data)
	x.iter.ReadVal(&val)
	return val, x.iter.Error
}

func (x *jsoniterBencher) valid(data []byte) bool {
	return jsoniter.Valid(data)
}

func (x *jsoniterBencher) readRepoData(data []byte, result *repoData) error {
	x.resetIter(data)
	iter := x.iter
	handler := jsoniterGetValuesFromRepoHelper{
		res: result,
	}
	iter.ReadObjectCB(handler.callback)
	return iter.Error
}

type jsoniterGetValuesFromRepoHelper struct {
	res          *repoData
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
