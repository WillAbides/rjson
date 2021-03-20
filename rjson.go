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

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\t': true,
	'\n': true,
}

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

type ObjectValueHandler interface {
	HandleObjectValue(fieldname, data []byte) (int, error)
}

type ObjectValueHandlerFunc func(fieldname, data []byte) (int, error)

func (fn ObjectValueHandlerFunc) HandleObjectValue(fieldname, data []byte) (int, error) {
	return fn(fieldname, data)
}

type ValueHandler interface {
	HandleValue(data []byte) (int, error)
}

type ValueHandlerFunc func(data []byte) (int, error)

func (fn ValueHandlerFunc) HandleValue(data []byte) (int, error) {
	return fn(data)
}

func ReadUint64(data []byte) (val uint64, n int, err error) {
	return readUint64(data)
}

func ReadInt64(data []byte) (val int64, n int, err error) {
	return readInt64(data)
}

func ReadInt32(data []byte) (val int32, n int, err error) {
	return readInt32(data)
}

func ReadUint32(data []byte) (val uint32, n int, err error) {
	return readUint32(data)
}

func ReadInt(data []byte) (val, n int, err error) {
	switch strconv.IntSize {
	case 64:
		val, n, err := ReadInt64(data)
		return int(val), n, err
	case 32:
		val, n, err := ReadInt32(data)
		return int(val), n, err
	default:
		return 0, 0, fmt.Errorf("unsupported int size: %d", strconv.IntSize)
	}
}

func ReadUint(data []byte) (val uint, n int, err error) {
	switch strconv.IntSize {
	case 64:
		val, n, err := ReadUint64(data)
		return uint(val), n, err
	case 32:
		val, n, err := ReadUint32(data)
		return uint(val), n, err
	default:
		return 0, 0, fmt.Errorf("unsupported int size: %d", strconv.IntSize)
	}
}

func ReadFloat64(data []byte) (val float64, n int, err error) {
	return readFloat64(data)
}

func ReadStringBytes(data, buf []byte) (val []byte, n int, err error) {
	p := countWhitespace(data)
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

func ReadString(data, buf []byte) (val string, n int, err error) {
	b, p, err := ReadStringBytes(data, buf)
	if err != nil {
		return "", p, err
	}
	return string(b), p, err
}

func ReadBool(data []byte) (val bool, n int, err error) {
	return readBool(data)
}

func ReadNull(data []byte) (n int, err error) {
	return readNull(data)
}

type Buffer struct {
	stackBuf []int
}

func (h *Buffer) HandleObjectValues(data []byte, handler ObjectValueHandler) (n int, err error) {
	n, h.stackBuf, err = handleObjectValues(data, handler, h.stackBuf)
	return n, err
}

func HandleObjectValues(data []byte, handler ObjectValueHandler) (int, error) {
	n, _, err := handleObjectValues(data, handler, nil)
	return n, err
}

func (h *Buffer) HandleArrayValues(data []byte, handler ValueHandler) (n int, err error) {
	n, h.stackBuf, err = handleArrayValues(data, handler, h.stackBuf)
	return n, err
}

func HandleArrayValues(data []byte, handler ValueHandler) (int, error) {
	n, _, err := handleArrayValues(data, handler, nil)
	return n, err
}

func (h *Buffer) SkipValue(data []byte) (n int, err error) {
	n, h.stackBuf, err = skipValue(data, h.stackBuf)
	return n, err
}

func SkipValue(data []byte) (int, error) {
	n, _, err := skipValue(data, nil)
	return n, err
}

func UnescapeStringContent(data, dst []byte) (val []byte, n int, err error) {
	return unescapeStringContent(data, dst)
}
