package rjson

import (
	"fmt"
	"math"
	"strconv"
	"sync"
)

//go:generate script/generate-ragel-file read_machines.rl

// ReadUint64 reads a uint64 value at the beginning of data. p is the first position in data after the value.
func ReadUint64(data []byte) (val uint64, p int, err error) {
	const cutoff = (1<<64-1)/10 + 1
	const zero = uint64('0')
	p = countWhitespace(data)
	if p == len(data) {
		return 0, p, errInvalidUInt
	}
	if data[p] == '0' {
		p++
		if len(data[p:]) == 0 {
			return 0, p, nil
		}
		switch data[p] {
		case '.', 'e', 'E':
			return 0, p, errInvalidUInt
		}
		return 0, p, nil
	}
	startP := p
	for ; p < len(data); p++ {
		if p-startP == 18 || data[p] < '0' || data[p] > '9' {
			break
		}
		val *= 10
		val += uint64(data[p]) - zero
	}
	if p-startP == 18 {
		for ; p < len(data); p++ {
			if data[p] < '0' || data[p] > '9' {
				break
			}
			if val > cutoff {
				return 0, p, fmt.Errorf(`value out of uint64 range`)
			}
			v := uint64(data[p]) - zero
			newVal := val * 10
			newVal += v
			if newVal < val {
				return 0, p, fmt.Errorf(`value out of uint64 range`)
			}
			val = newVal
		}
	}
	if p-startP == 0 {
		return 0, p, errInvalidUInt
	}
	if p == len(data) {
		return val, p, nil
	}
	switch data[p] {
	case '.', 'e', 'E':
		return 0, p, errInvalidUInt
	}
	return val, p, nil
}

// ReadUint32 reads a uint32 value at the beginning of data. p is the first position in data after the value.
func ReadUint32(data []byte) (val uint32, p int, err error) {
	var val64 uint64
	val64, p, err = ReadUint64(data)
	if err == nil && val64 > math.MaxUint32 {
		val64 = 0
		err = errInvalidUInt
	}
	return uint32(val64), p, err
}

// ReadInt64 reads an int64 value at the beginning of data. p is the first position in data after the value.
func ReadInt64(data []byte) (val int64, p int, err error) {
	const cutoff = uint64(1 << uint64(63))
	p = countWhitespace(data)
	if p == len(data) {
		return 0, p, errInvalidInt
	}
	neg := data[0] == '-'
	if neg {
		p++
		if p == len(data) {
			return 0, p, errInvalidInt
		}
	}
	var u64Val uint64
	var pp int
	u64Val, pp, err = ReadUint64(data[p:])
	p += pp
	if err != nil {
		return 0, p, err
	}
	if neg {
		if u64Val > cutoff {
			return 0, p, fmt.Errorf("value out of int64 range")
		}
		return -int64(u64Val), p, nil
	}
	if u64Val >= cutoff {
		return 0, p, fmt.Errorf("value out of int64 range")
	}
	return int64(u64Val), p, nil
}

// ReadInt32 reads an int32 value at the beginning of data. p is the first position in data after the value.
func ReadInt32(data []byte) (val int32, p int, err error) {
	var val64 int64
	val64, p, err = ReadInt64(data)
	if err != nil {
		return 0, p, err
	}
	if val64 > math.MaxInt32 || val64 < math.MinInt32 {
		return 0, p, errInvalidInt
	}
	return int32(val64), p, nil
}

// ReadInt reads an int value at the beginning of data. p is the first position in data after the value.
func ReadInt(data []byte) (val, p int, err error) {
	switch strconv.IntSize {
	case 64:
		val, p, err := ReadInt64(data)
		return int(val), p, err
	case 32:
		val, p, err := ReadInt32(data)
		return int(val), p, err
	default:
		return 0, 0, fmt.Errorf("unsupported int size: %d", strconv.IntSize)
	}
}

// ReadUint reads a uint value at the beginning of data. p is the first position in data after the value.
func ReadUint(data []byte) (val uint, p int, err error) {
	switch strconv.IntSize {
	case 64:
		val, p, err := ReadUint64(data)
		return uint(val), p, err
	case 32:
		val, p, err := ReadUint32(data)
		return uint(val), p, err
	default:
		return 0, 0, fmt.Errorf("unsupported int size: %d", strconv.IntSize)
	}
}

// ReadFloat64 reads a float64 value at the beginning of data. p is the first position in data after the value.
func ReadFloat64(data []byte) (val float64, p int, err error) {
	p = countWhitespace(data)
	if p == len(data) {
		return 0, p, errInvalidNumber
	}
	return readFloat64(data[p:])
}

// ReadStringBytes reads a string value at the beginning of data and appends it to buf. p is the first position in
// data after the value.
func ReadStringBytes(data, buf []byte) (val []byte, p int, err error) {
	p = countWhitespace(data)
	if p == len(data) || data[p] != '"' {
		return buf, p, fmt.Errorf("not a string")
	}
	p++
	start := p
	for ; p < len(data); p++ {
		var pp int
		var err error
		if data[p] <= 0x1f {
			buf, pp, err = appendRemainderOfString(data[p:], buf)
			p += pp
			return buf, p, err
		}
		switch data[p] {
		case '"':
			return append(buf, data[start:p]...), p + 1, nil
		case '\\':
			buf = append(buf, data[start:p]...)
			buf, pp, err = appendRemainderOfString(data[p:], buf)
			p += pp
			return buf, p, err
		}
	}
	return buf, p, fmt.Errorf("not a string")
}

// ReadString reads a string value at the beginning of data. p is the first position in data after the value. If buf
// is not nil, it will be used as a working for building the string value.
//
// If you are concerned about memory allocation, try using ReadStringBytes instead.
func ReadString(data, buf []byte) (val string, p int, err error) {
	buf = buf[:0]
	p = countWhitespace(data)
	if p == len(data) || data[p] != '"' {
		return "", p, fmt.Errorf("not a string")
	}
	p++
	start := p

	for ; p < len(data); p++ {
		var pp int
		var err error
		if data[p] <= 0x1f {
			buf, pp, err = appendRemainderOfString(data[p:], buf)
			p += pp
			return string(buf), p, err
		}
		switch data[p] {
		case '"':
			return string(data[start:p]), p + 1, nil
		case '\\':
			buf = append(buf, data[start:p]...)
			buf, pp, err = appendRemainderOfString(data[p:], buf)
			p += pp
			return string(buf), p, err
		}
	}
	return "", p, fmt.Errorf("not a string")
}

// ReadBool reads a boolean value at the beginning of data. p is the first position in data after the value.
func ReadBool(data []byte) (val bool, p int, err error) {
	return readBool(data)
}

// ReadNull reads 'null' at the beginning of data. p is the first position after 'null'
func ReadNull(data []byte) (p int, err error) {
	return readNull(data)
}

// ValueReader is a handler for reading complex json data types (Objects and Arrays).
type ValueReader struct {
	buf          Buffer
	pool         sync.Pool
	objVal       map[string]interface{}
	arrVal       []interface{}
	fieldNameBuf []byte
	stringBuf    []byte

	newMapSize  int
	lastMapSize int
	maxMapSize  int

	newSliceSize  int
	lastSliceSize int
}

func (h *ValueReader) borrowValueReader() *ValueReader {
	x, ok := h.pool.Get().(*ValueReader)
	if !ok {
		x = &ValueReader{}
	}
	x.newMapSize = 0
	return x
}

func (h *ValueReader) returnValueReader(x *ValueReader) {
	x.arrVal = x.arrVal[:0]
	h.pool.Put(x)
}

// HandleValue implements ValueHandler.HandleValue
func (h *ValueReader) HandleValue(data []byte) (p int, err error) {
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
		h2.newMapSize = h.maxMapSize
		val, pp, err = h2.ReadObject(data)
		mpLen := len(val.(map[string]interface{}))
		if mpLen > h.maxMapSize {
			h.maxMapSize = mpLen
		}
		h.returnValueReader(h2)
	case ArrayStartType:
		h2 := h.borrowValueReader()
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
			h.fieldNameBuf, _, err = unescapeStringContent(fieldname[i:], append(h.fieldNameBuf[:0], fieldname[:i]...))
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
		h2.newMapSize = h.maxMapSize
		val, pp, err = h2.ReadObject(data)
		mpLen := len(val.(map[string]interface{}))
		if mpLen > h.maxMapSize {
			h.maxMapSize = mpLen
		}
		h.returnValueReader(h2)
	case ArrayStartType:
		h2 := h.borrowValueReader()
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

// ReadObject allocates a ValueReader and returns ValueReader.ReadObject. You should probably use ValueReader.ReadObject
// instead so you don't have to allocate a new ValueReader for each call.
func ReadObject(data []byte) (val map[string]interface{}, p int, err error) {
	h := ValueReader{}
	return h.ReadObject(data)
}

// ReadObject reads an object value from the front of data and returns it as a map[string]interface{}. p is the first
// position in data after the value.
func (h *ValueReader) ReadObject(data []byte) (val map[string]interface{}, p int, err error) {
	mapSize := h.newMapSize
	if mapSize == 0 {
		mapSize = h.lastMapSize
	}
	h.objVal = make(map[string]interface{}, mapSize)
	p, err = h.buf.HandleObjectValues(data[p:], h)
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

// ReadArray allocates a ValueReader and returns ValueReader.ReadArray. You should probably use ValueReader.ReadArray
// instead so you don't have to allocate a new ValueReader for each call.
func ReadArray(data []byte) (val []interface{}, p int, err error) {
	h := ValueReader{}
	return h.ReadArray(data)
}

// ReadArray reads an array from the front of data and returns it as a []interface{}. p is the first position in data
// after the value.
func (h *ValueReader) ReadArray(data []byte) (val []interface{}, p int, err error) {
	sliceSize := h.newSliceSize
	if sliceSize == 0 {
		sliceSize = h.lastSliceSize
	}
	h.arrVal = make([]interface{}, 0, sliceSize)
	p, err = h.buf.HandleArrayValues(data, h)
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
