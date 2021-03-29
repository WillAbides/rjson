package rjson

import (
	"io"
)

//go:generate script/generate-ragel-file misc_machines.rl
//go:generate script/generate-ragel-file skip_machine.rl
//go:generate script/generate-ragel-file array_handler_machine.rl
//go:generate script/generate-ragel-file object_handler_machine.rl

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

// SkipValueFast is like SkipValue but it speeds things up by skipping validation on objects and arrays.
func (h *Buffer) SkipValueFast(data []byte) (p int, err error) {
	p, h.stackBuf, err = skipValueFast(data, h.stackBuf)
	return p, err
}

// SkipValueFast is like SkipValue but it speeds things up by skipping validation on objects and arrays.
func SkipValueFast(data []byte) (p int, err error) {
	p, _, err = skipValueFast(data, nil)
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
	p, h.stackBuf, err = skipValue(data, h.stackBuf)
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
