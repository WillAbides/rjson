package benchmarks

import (
	"fmt"
	"io"

	jsoniter "github.com/json-iterator/go"
)

type jsoniterBencher struct {
	iter                          *jsoniter.Iterator
	jsoniterDistinctUerIdsHandler *jsoniterDistinctUerIdsHandler
}

func (x *jsoniterBencher) name() string {
	return "jsoniter"
}

func (x *jsoniterBencher) init() {
	*x = jsoniterBencher{
		iter:                          jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary),
		jsoniterDistinctUerIdsHandler: &jsoniterDistinctUerIdsHandler{},
	}
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

func (x *jsoniterBencher) decodeInt64(data []byte, v *int64) error {
	return jsoniter.Unmarshal(data, v)
}

func (x *jsoniterBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	x.resetIter(data)
	x.iter.ReadVal(&val)
	return val, x.iter.Error
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

type jsoniterDistinctUerIdsHandler struct {
	userIDs  []int64
	inUser   bool
	inStatus bool
}

func (x *jsoniterBencher) distinctUserIDs(data []byte, dest []int64) ([]int64, error) {
	x.resetIter(data)
	x.jsoniterDistinctUerIdsHandler.userIDs = dest
	x.iter.ReadObjectCB(x.jsoniterDistinctUerIdsHandler.objectCB)
	if x.iter.Error != nil {
		return nil, x.iter.Error
	}
	return x.jsoniterDistinctUerIdsHandler.userIDs, nil
}

func (h *jsoniterDistinctUerIdsHandler) objectCB(it *jsoniter.Iterator, fieldname string) bool {
	switch fieldname {
	case "statuses":
		return it.ReadArrayCB(h.arrayCB)
	case "user":
		if h.inUser || !h.inStatus {
			it.Skip()
			return true
		}
		h.inUser = true
		it.ReadObjectCB(h.objectCB)
		h.inUser = false
		return true
	case "id":
		if !h.inUser {
			it.Skip()
			return true
		}
		h.userIDs = append(h.userIDs, it.ReadInt64())
		return true
	}
	it.Skip()
	return true
}

func (h *jsoniterDistinctUerIdsHandler) arrayCB(it *jsoniter.Iterator) bool {
	h.inStatus = true
	it.ReadObjectCB(h.objectCB)
	h.inStatus = false
	return true
}

func (x *jsoniterBencher) readString(data []byte) (string, error) {
	x.resetIter(data)
	if x.iter.WhatIsNext() != jsoniter.StringValue {
		return "", fmt.Errorf("not a string")
	}
	s := x.iter.ReadString()
	if x.iter.Error != nil {
		return "", x.iter.Error
	}
	return s, nil
}

func (x *jsoniterBencher) readBool(data []byte) (bool, error) {
	x.resetIter(data)
	if x.iter.WhatIsNext() != jsoniter.BoolValue {
		return false, fmt.Errorf("not a bool")
	}
	return x.iter.ReadBool(), x.iter.Error
}
