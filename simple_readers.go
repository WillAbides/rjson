package rjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/willabides/rjson/internal/fp"
)

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

func readUint64Compat(data []byte) (val uint64, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
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

func readUint32Compat(data []byte) (val uint32, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
}

// ReadInt64 reads an int64 value at the beginning of data. p is the first position in data after the value.
func ReadInt64(data []byte) (val int64, p int, err error) {
	const cutoff = uint64(1 << uint64(63))
	p = countWhitespace(data)
	if p == len(data) {
		return 0, p, errInvalidInt
	}
	neg := data[p] == '-'
	if neg {
		p++
		if p == len(data) || whitespace[data[p]] {
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

func readInt64Compat(data []byte) (val int64, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
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

func readInt32Compat(data []byte) (val int32, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
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

func readIntCompat(data []byte) (val, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
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

func readUintCompat(data []byte) (val uint, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
}

// ReadFloat64 reads a float64 value at the beginning of data. p is the first position in data after the value.
func ReadFloat64(data []byte) (val float64, p int, err error) {
	p = countWhitespace(data)
	if p == len(data) {
		return 0, p, errInvalidNumber
	}
	var pp int
	val, pp, err = fp.ParseJSONFloatPrefix(data[p:])
	return val, p + pp, err
}

func readFloat64Compat(data []byte) (val float64, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	_, ok := token.(json.Number)
	if !ok {
		return 0, 0, errInvalidNumber
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
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

func readStringBytesCompat(data []byte) (val []byte, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return nil, 0, err
	}
	_, ok := token.(string)
	if !ok {
		return nil, 0, errInvalidString
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	var strVal string
	err = decoder.Decode(&strVal)
	return []byte(strVal), int(decoder.InputOffset()), err
}

// ReadString reads a string value at the beginning of data. p is the first position in data after the value. If buf
// is not nil, it will be used as a working buffer for building the string value.
//
// If you are concerned about memory allocation, try using ReadStringBytes instead.
func ReadString(data []byte, buf *[]byte) (val string, p int, err error) {
	p = countWhitespace(data)
	if p == len(data) || data[p] != '"' {
		return "", p, fmt.Errorf("not a string")
	}
	p++
	start := p
	var bBuf []byte
	if buf != nil {
		bBuf = (*buf)[:0]
	}
	for ; p < len(data); p++ {
		var pp int
		var err error
		if data[p] <= 0x1f {
			bBuf, pp, err = appendRemainderOfString(data[p:], bBuf)
			p += pp
			if buf != nil {
				*buf = bBuf
			}
			return string(bBuf), p, err
		}
		switch data[p] {
		case '"':
			return string(data[start:p]), p + 1, nil
		case '\\':
			bBuf = append(bBuf, data[start:p]...)
			bBuf, pp, err = appendRemainderOfString(data[p:], bBuf)
			p += pp
			if buf != nil {
				*buf = bBuf
			}
			return string(bBuf), p, err
		}
	}
	return "", p, fmt.Errorf("not a string")
}

func readStringCompat(data []byte) (val string, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return "", 0, err
	}
	_, ok := token.(string)
	if !ok {
		return "", 0, errInvalidString
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&val)
	return val, int(decoder.InputOffset()), err
}

// ReadBool reads a boolean value at the beginning of data. p is the first position in data after the value.
func ReadBool(data []byte) (val bool, p int, err error) {
	return readBool(data)
}

func readBoolCompat(data []byte) (val bool, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return false, 0, err
	}
	var ok bool
	val, ok = token.(bool)
	if !ok {
		return false, 0, errNotBool
	}
	return val, int(decoder.InputOffset()), nil
}

// ReadNull reads 'null' at the beginning of data. p is the first position after 'null'
func ReadNull(data []byte) (p int, err error) {
	return readNull(data)
}

func readNullCompat(data []byte) (p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var token json.Token
	token, err = decoder.Token()
	if err != nil {
		return 0, errNotNull
	}
	if token != nil {
		return 0, errNotNull
	}
	return int(decoder.InputOffset()), nil
}
