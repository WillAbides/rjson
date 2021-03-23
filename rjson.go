package rjson

import (
	"fmt"
	"io"
	"strconv"
)

//go:generate script/generate-ragel-file misc_machines.rl
//go:generate script/generate-ragel-file skip_machine.rl
//go:generate script/generate-ragel-file array_handler_machine.rl
//go:generate script/generate-ragel-file object_handler_machine.rl
//go:generate script/generate-ragel-file read_machines.rl

// NextToken finds the first json token in data. token is the token itself, p is the position in data where
// the token was found. NextToken errors if it finds anything besides json whitespace before the first valid
// token. It returns io.EOF if data is empty or contains only whitespace.
func NextToken(data []byte) (token byte, p int, err error) {
	if len(data) == 0 {
		return 0, 0, io.EOF
	}
	if tokenTypes[data[0]] != InvalidType {
		return data[0], 1, nil
	}
	if !whitespace[data[0]] {
		return data[0], 1, errNoValidToken
	}
	p = countWhitespace(data)
	if p >= len(data) {
		return 0, p, io.EOF
	}
	b := data[p]
	if tokenTypes[b] == InvalidType {
		return b, p + 1, errNoValidToken
	}
	return b, p + 1, nil
}

// NextTokenType finds the first json token in data and returns its TokenType. p is the position in data where
// the token was found. NextToken errors if it finds anything besides json whitespace before the first valid
// token. It returns io.EOF if data is empty or contains only whitespace.
func NextTokenType(data []byte) (TokenType, int, error) {
	if len(data) == 0 {
		return 0, 0, io.EOF
	}
	tp := tokenTypes[data[0]]
	if tp != InvalidType {
		return tp, 1, nil
	}
	if !whitespace[data[0]] {
		return tp, 1, nil
	}
	p := countWhitespace(data)
	if p >= len(data) {
		return 0, p, io.EOF
	}
	return tokenTypes[data[p]], p + 1, nil
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\t': true,
	'\n': true,
}

// ObjectValueHandler is a handler for json objects.
type ObjectValueHandler interface {
	HandleObjectValue(fieldname, data []byte) (p int, err error)
}

// ObjectValueHandlerFunc is a function that implements ObjectValueHandler
type ObjectValueHandlerFunc func(fieldname, data []byte) (p int, err error)

// HandleObjectValue meets ObjectValueHandler.HandleObjectValue
func (fn ObjectValueHandlerFunc) HandleObjectValue(fieldname, data []byte) (int, error) {
	return fn(fieldname, data)
}

// ValueHandler is a handler for json values
type ValueHandler interface {
	HandleValue(data []byte) (p int, err error)
}

// ValueHandlerFunc is a function that implements ValueHandler
type ValueHandlerFunc func(data []byte) (p int, err error)

// HandleValue meets ValueHandler.HandleValue
func (fn ValueHandlerFunc) HandleValue(data []byte) (int, error) {
	return fn(data)
}

// ReadUint64 reads a uint64 value at the beginning of data. p is the first position in data after the value.
func ReadUint64(data []byte) (val uint64, p int, err error) {
	return readUint64(data)
}

// ReadInt64 reads an int64 value at the beginning of data. p is the first position in data after the value.
func ReadInt64(data []byte) (val int64, p int, err error) {
	return readInt64(data)
}

// ReadInt32 reads an int32 value at the beginning of data. p is the first position in data after the value.
func ReadInt32(data []byte) (val int32, p int, err error) {
	return readInt32(data)
}

// ReadUint32 reads a uint32 value at the beginning of data. p is the first position in data after the value.
func ReadUint32(data []byte) (val uint32, p int, err error) {
	return readUint32(data)
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
	return readFloat64(data)
}

// ReadStringBytes reads a string value at the beginning of data and appends it to buf. p is the first position in
// data after the value.
func ReadStringBytes(data, buf []byte) (val []byte, p int, err error) {
	p = countWhitespace(data)
	if data[p] != '"' {
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
	return data, p, fmt.Errorf("not a string")
}

// ReadString reads a string value at the beginning of data. p is the first position in data after the value. If buf
// is not nil, it will be used as a working for building the string value.
//
// If you are concerned about memory allocation, try using ReadStringBytes instead.
func ReadString(data, buf []byte) (val string, p int, err error) {
	b, p, err := ReadStringBytes(data, buf)
	if err != nil {
		return "", p, err
	}
	return string(b), p, err
}

// ReadBool reads a boolean value at the beginning of data. p is the first position in data after the value.
func ReadBool(data []byte) (val bool, p int, err error) {
	return readBool(data)
}

// ReadNull reads 'null' at the beginning of data. p is the first position after 'null'
func ReadNull(data []byte) (p int, err error) {
	return readNull(data)
}

// Buffer is a reusable stack buffer for operations that walk a json document.
type Buffer struct {
	stackBuf []int
}

// HandleObjectValues runs handler.HandleObjectValue on each field in the object at the beginning of data until it
// encounters an error or reaches the end of the object. p is the position after the last byte it read. When err
// is nil, p will be the position after the object.
func (h *Buffer) HandleObjectValues(data []byte, handler ObjectValueHandler) (p int, err error) {
	p, h.stackBuf, err = handleObjectValues(data, handler, h.stackBuf)
	return p, err
}

// HandleObjectValues runs handler.HandleObjectValue on each field in the object at the beginning of data until it
// encounters an error or reaches the end of the object. p is the position after the last byte it read. When err
// is nil, p will be the position after the object.
func HandleObjectValues(data []byte, handler ObjectValueHandler) (p int, err error) {
	p, _, err = handleObjectValues(data, handler, nil)
	return p, err
}

// HandleArrayValues runs handler.HandleValue on each item in the array at the beginning of data until it encounters
// an error or reaches the end of the array. p is the position after the last byte it read. When err is nil, p will
// be the position after the object.
func (h *Buffer) HandleArrayValues(data []byte, handler ValueHandler) (p int, err error) {
	p, h.stackBuf, err = handleArrayValues(data, handler, h.stackBuf)
	return p, err
}

// HandleArrayValues runs handler.HandleValue on each item in the array at the beginning of data until it encounters
// an error or reaches the end of the array. p is the position after the last byte it read. When err is nil, p will
// be the position after the object.
func HandleArrayValues(data []byte, handler ValueHandler) (p int, err error) {
	n, _, err := handleArrayValues(data, handler, nil)
	return n, err
}

// SkipValue skips the first json value in data. p is the position after the skipped value.
func (h *Buffer) SkipValue(data []byte) (p int, err error) {
	p, h.stackBuf, err = skipValue(data, h.stackBuf)
	return p, err
}

// SkipValue skips the first json value in data. p is the position after the skipped value.
func SkipValue(data []byte) (p int, err error) {
	p, _, err = skipValue(data, nil)
	return p, err
}

// UnescapeStringContent unescapes the content of a raw json string. data must be the content of the raw string without
// starting and ending double quotes. This function is useful with ObjectValueHandler.HandleObjectValue.
func UnescapeStringContent(data, dst []byte) (val []byte, p int, err error) {
	return unescapeStringContent(data, dst)
}

// Valid returns true if data contains a single valid json value
func (h *Buffer) Valid(data []byte) bool {
	var p int
	var err error
	_, h.stackBuf, err = skipValue(data, h.stackBuf)
	if err != nil {
		return false
	}
	if p > len(data) {
		return true
	}
	return p+countWhitespace(data[p:]) >= len(data)
}

// Valid returns true if data contains a single valid json value
func Valid(data []byte) bool {
	var p int
	var err error
	p, _, err = skipValue(data, nil)
	if err != nil {
		return false
	}
	if p > len(data) {
		return true
	}
	return p+countWhitespace(data[p:]) >= len(data)
}
