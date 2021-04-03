package rjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

//go:generate script/generate-ragel-file read_machines.rl

const valueReaderMaxDepth = 10_000

// ValueReader is a handler for reading complex json data types (Objects and Arrays).
type ValueReader struct {
	buf          Buffer
	pool         sync.Pool
	objVal       map[string]interface{}
	arrVal       []interface{}
	fieldNameBuf []byte
	stringBuf    []byte
	depth        int

	newMapSize  int
	lastMapSize int
	maxMapSize  int

	newSliceSize  int
	lastSliceSize int
}

func (h *ValueReader) borrowValueReader() *ValueReader {
	x, ok := h.pool.Get().(*ValueReader)
	if !ok {
		x = &ValueReader{
			depth: h.depth + 1,
		}
	}
	x.newMapSize = 0
	x.depth = h.depth + 1
	return x
}

func (h *ValueReader) returnValueReader(x *ValueReader) {
	x.arrVal = x.arrVal[:0]
	h.pool.Put(x)
}

// HandleArrayValue implements ArrayValueHandler.HandleArrayValue
func (h *ValueReader) HandleArrayValue(data []byte) (p int, err error) {
	var tknType TokenType
	tknType, p, err = NextTokenType(data)
	if err != nil {
		return p, err
	}
	p--
	data = data[p:]
	var val interface{}
	var pp int
	switch tknType {
	case ObjectStartType:
		h2 := h.borrowValueReader()
		if h2.depth > valueReaderMaxDepth {
			return p, errMaxDepth
		}
		h2.newMapSize = h.maxMapSize
		val, pp, err = h2.ReadObject(data)
		mpLen := len(val.(map[string]interface{}))
		if mpLen > h.maxMapSize {
			h.maxMapSize = mpLen
		}
		h.returnValueReader(h2)
	case ArrayStartType:
		h2 := h.borrowValueReader()
		if h2.depth > valueReaderMaxDepth {
			return p, errMaxDepth
		}
		val, pp, err = h2.ReadArray(data)
		h.returnValueReader(h2)
	default:
		val, pp, err = h.readSimpleValue(data, tknType)
	}
	if err == nil {
		h.arrVal = append(h.arrVal, val)
	}
	return p + pp, err
}

// HandleObjectValue implements ValueObjectHandler.HandleObjectValue.
func (h *ValueReader) HandleObjectValue(fieldname, data []byte) (p int, err error) {
	for i := 0; i < len(fieldname); i++ {
		if fieldname[i] == '\\' {
			h.fieldNameBuf, _, err = UnescapeStringContent(fieldname[i:], append(h.fieldNameBuf[:0], fieldname[:i]...))
			if err != nil {
				return 0, err
			}
			fieldname = h.fieldNameBuf
			break
		}
	}

	var tknType TokenType
	tknType, p, err = NextTokenType(data)
	if err != nil {
		return p, err
	}
	p--
	data = data[p:]
	var val interface{}
	var pp int
	switch tknType {
	case ObjectStartType:
		h2 := h.borrowValueReader()
		if h2.depth > valueReaderMaxDepth {
			return p, errMaxDepth
		}
		h2.newMapSize = h.maxMapSize
		val, pp, err = h2.ReadObject(data)
		mpLen := len(val.(map[string]interface{}))
		if mpLen > h.maxMapSize {
			h.maxMapSize = mpLen
		}
		h.returnValueReader(h2)
	case ArrayStartType:
		h2 := h.borrowValueReader()
		if h2.depth > valueReaderMaxDepth {
			return p, errMaxDepth
		}
		val, pp, err = h2.ReadArray(data)
		h.returnValueReader(h2)
	default:
		val, pp, err = h.readSimpleValue(data, tknType)
	}
	h.objVal[string(fieldname)] = val
	return p + pp, err
}

func (h *ValueReader) readSimpleValue(data []byte, tknType TokenType) (val interface{}, p int, err error) {
	switch tknType {
	case NullType:
		p, err = ReadNull(data)
		return nil, p, err
	case StringType:
		h.stringBuf, p, err = ReadStringBytes(data, h.stringBuf[:0])
		return string(h.stringBuf), p, err
	case NumberType:
		return ReadFloat64(data)
	case TrueType, FalseType:
		return ReadBool(data)
	default:
		return nil, p, fmt.Errorf("unexpected type")
	}
}

// ReadValue allocates a ValueReader and returns ValueReader.ReadValue. You should probably use ValueReader.ReadValue
// instead so you don't have to allocate a new ValueReader for each call.
func ReadValue(data []byte) (val interface{}, p int, err error) {
	h := ValueReader{}
	return h.ReadValue(data)
}

// ReadValue reads a value at the beginning of data. The result will be a string, bool, nil, float64, []interface{}
// or map[string]interface{} depending on the json data type. p is the first position in data after the value.
func (h *ValueReader) ReadValue(data []byte) (val interface{}, p int, err error) {
	var tknType TokenType
	tknType, p, err = NextTokenType(data)
	if err != nil {
		return nil, p, err
	}
	p--
	data = data[p:]

	var pp int
	switch tknType {
	case ObjectStartType:
		h2 := h.borrowValueReader()
		val, pp, err = h2.ReadObject(data)
		h.returnValueReader(h2)
	case ArrayStartType:
		h2 := h.borrowValueReader()
		val, pp, err = h2.ReadArray(data)
		h.returnValueReader(h2)
	default:
		val, pp, err = h.readSimpleValue(data, tknType)
	}
	if err != nil {
		return nil, p + pp, err
	}
	return val, p + pp, err
}

func readValueCompat(data []byte) (val interface{}, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	if err != nil {
		return nil, int(decoder.InputOffset()), fmt.Errorf("invalid json")
	}
	return val, int(decoder.InputOffset()), nil
}

// ReadObject allocates a ValueReader and returns ValueReader.ReadObject. You should probably use ValueReader.ReadObject
// instead so you don't have to allocate a new ValueReader for each call.
func ReadObject(data []byte) (val map[string]interface{}, p int, err error) {
	h := ValueReader{}
	return h.ReadObject(data)
}

// ReadObject reads an object value from the front of data and returns it as a map[string]interface{}. p is the first
// position in data after the value.
func (h *ValueReader) ReadObject(data []byte) (val map[string]interface{}, p int, err error) {
	if h.depth == 0 {
		h.depth = 1
		defer func() { h.depth = 0 }()
	}
	mapSize := h.newMapSize
	if mapSize == 0 {
		mapSize = h.lastMapSize
	}
	h.objVal = make(map[string]interface{}, mapSize)
	p, err = HandleObjectValues(data[p:], h, &h.buf)
	if err != nil {
		return nil, p, err
	}
	valLen := len(h.objVal)

	// make sure to return err for null
	if valLen == 0 {
		tknType, _, tknErr := NextTokenType(data)
		if tknErr == nil && tknType == NullType {
			return nil, p, errInvalidObject
		}
	}

	h.lastMapSize = valLen
	return h.objVal, p, nil
}

func readObjectCompat(data []byte) (val map[string]interface{}, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	tkn, err := decoder.Token()
	if err != nil {
		return nil, int(decoder.InputOffset()), errInvalidObject
	}
	if tkn != json.Delim('{') {
		return nil, int(decoder.InputOffset()), errInvalidObject
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	if err != nil {
		return nil, int(decoder.InputOffset()), err
	}
	return val, int(decoder.InputOffset()), nil
}

// ReadArray allocates a ValueReader and returns ValueReader.ReadArray. You should probably use ValueReader.ReadArray
// instead so you don't have to allocate a new ValueReader for each call.
func ReadArray(data []byte) (val []interface{}, p int, err error) {
	h := ValueReader{}
	return h.ReadArray(data)
}

// ReadArray reads an array from the front of data and returns it as a []interface{}. p is the first position in data
// after the value.
func (h *ValueReader) ReadArray(data []byte) (val []interface{}, p int, err error) {
	if h.depth == 0 {
		h.depth = 1
		defer func() {
			h.depth = 0
		}()
	}
	sliceSize := h.newSliceSize
	if sliceSize == 0 {
		sliceSize = h.lastSliceSize
	}
	h.arrVal = make([]interface{}, 0, sliceSize)
	p, err = HandleArrayValues(data, h, &h.buf)
	if err != nil {
		return nil, p, err
	}

	valLen := len(h.arrVal)

	// make sure to return err for null
	if valLen == 0 {
		tknType, _, tknErr := NextTokenType(data)
		if tknErr == nil && tknType == NullType {
			return nil, p, errInvalidArray
		}
	}

	h.lastSliceSize = valLen
	return h.arrVal, p, err
}

func readArrayCompat(data []byte) (val []interface{}, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	tkn, err := decoder.Token()
	if err != nil {
		return nil, int(decoder.InputOffset()), errInvalidArray
	}
	if tkn != json.Delim('[') {
		return nil, int(decoder.InputOffset()), errInvalidArray
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	if err != nil {
		return nil, int(decoder.InputOffset()), err
	}
	return val, int(decoder.InputOffset()), nil
}
