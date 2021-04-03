package rjson

import (
	"io"
	"unicode/utf8"
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

// ObjectValueHandler is a handler for json objects.
type ObjectValueHandler interface {
	HandleObjectValue(fieldname, data []byte) (p int, err error)
}

// ObjectValueHandlerFunc is a function that implements ObjectValueHandler
type ObjectValueHandlerFunc func(fieldname, data []byte) (p int, err error)

// HandleObjectValue implements ObjectValueHandler.HandleObjectValue
func (fn ObjectValueHandlerFunc) HandleObjectValue(fieldname, data []byte) (int, error) {
	return fn(fieldname, data)
}

// ArrayValueHandler is a handler for values in a json array
type ArrayValueHandler interface {
	HandleArrayValue(data []byte) (p int, err error)
}

// ArrayValueHandlerFunc is a function that implements ArrayValueHandler
type ArrayValueHandlerFunc func(data []byte) (p int, err error)

// HandleArrayValue implements ArrayValueHandler.HandleArrayValue
func (fn ArrayValueHandlerFunc) HandleArrayValue(data []byte) (int, error) {
	return fn(data)
}

// Buffer is a reusable stack buffer for functions that read nested objects and arrays.
// Buffer is not thread-safe.
type Buffer struct {
	stackBuf []int
}

// HandleObjectValues runs handler.HandleObjectValue on each field in the object at the beginning of data until it
// encounters an error or reaches the end of the object. p is the position after the last byte it read. When err
// is nil, p will be the position after the object.
// buffer is optional. Reusing a buffer can reduce memory allocations.
func HandleObjectValues(data []byte, handler ObjectValueHandler, buffer *Buffer) (p int, err error) {
	if buffer == nil {
		p, _, err = handleObjectValues(data, handler, nil)
		return p, err
	}
	p, buffer.stackBuf, err = handleObjectValues(data, handler, buffer.stackBuf)
	return p, err
}

// HandleArrayValues runs handler.HandleArrayValue on each item in the array at the beginning of data until it encounters
// an error or reaches the end of the array. p is the position after the last byte it read. When err is nil, p will
// be the position after the object.
// buffer is optional. Reusing a buffer can reduce memory allocations.
func HandleArrayValues(data []byte, handler ArrayValueHandler, buffer *Buffer) (p int, err error) {
	if buffer == nil {
		p, _, err = handleArrayValues(data, handler, nil)
		return p, err
	}
	p, buffer.stackBuf, err = handleArrayValues(data, handler, buffer.stackBuf)
	return p, err
}

// SkipValue skips the first json value in data. p is the position after the skipped value.
// buffer is optional. Reusing a buffer can reduce memory allocations.
func SkipValue(data []byte, buffer *Buffer) (p int, err error) {
	if buffer == nil {
		p, _, err = skipValue(data, nil)
		return p, err
	}
	p, buffer.stackBuf, err = skipValue(data, buffer.stackBuf)
	return p, err
}

// SkipValueFast is like SkipValue but it speeds things up by skipping validation on objects and arrays.
// buffer is optional. Reusing a buffer can reduce memory allocations.
func SkipValueFast(data []byte, buffer *Buffer) (p int, err error) {
	if buffer == nil {
		p, _, err = skipValueFast(data, nil)
		return p, err
	}
	p, buffer.stackBuf, err = skipValueFast(data, buffer.stackBuf)
	return p, err
}

// UnescapeStringContent unescapes the content of a raw json string. data must be the content of the raw string without
// starting and ending double quotes. This function is useful with ObjectValueHandler.HandleObjectValue.
func UnescapeStringContent(data, dst []byte) (val []byte, p int, err error) {
	return unescapeStringContent(data, dst)
}

// Valid returns true if data contains a single valid json value.
// buffer is optional. Reusing a buffer can reduce memory allocations.
func Valid(data []byte, buffer *Buffer) bool {
	var p int
	var err error
	if buffer == nil {
		p, _, err = skipValue(data, nil)
	} else {
		p, buffer.stackBuf, err = skipValue(data, buffer.stackBuf)
	}

	if err != nil {
		return false
	}
	if p > len(data) {
		return true
	}
	return p+countWhitespace(data[p:]) >= len(data)
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\t': true,
	'\n': true,
}

func countWhitespace(data []byte) int {
	for i := 0; i < len(data); i++ {
		if !whitespace[data[i]] {
			return i
		}
	}
	return len(data)
}

// StdLibCompatibleString replaces invalid utf8 with utf8.RuneError to match the standard library's behavior.
func StdLibCompatibleString(rjsonString string) string {
	runes := make([]rune, 0, len(rjsonString))
	var i int
	for i < len(rjsonString) {
		r, w := utf8.DecodeRuneInString(rjsonString[i:])
		i+=w
		runes = append(runes, r)
	}
	return string(runes)
}

// StdLibCompatibleStringBytes replaces invalid utf8 with utf8.RuneError to match the standard library's behavior.
func StdLibCompatibleStringBytes(rjsonString []byte, buf []byte) []byte {
	var i int
	for i < len(rjsonString) {
		r, w := utf8.DecodeRune(rjsonString[i:])
		if cap(buf) - len(buf) < w {
			buf = growBytesSliceCapacity(buf, w + (len(rjsonString) - i) * utf8.UTFMax)
		}
		i += w
		buf = append(buf, string(r)...)
	}
	return buf
}

// StdLibCompatibleSlice returns a copy of rjsonSlice with StdLibCompatibleString applied to all string values recursively.
func StdLibCompatibleSlice(rjsonSlice []interface{}) []interface{} {
	out := make([]interface{}, len(rjsonSlice))
	for i, val := range rjsonSlice {
		switch v := val.(type) {
		case string:
			out[i] = StdLibCompatibleString(v)
		case []interface{}:
			out[i] = StdLibCompatibleSlice(v)
		case map[string]interface{}:
			out[i] = StdLibCompatibleMap(v)
		default:
			out[i] = v
		}
	}
	return out
}

// StdLibCompatibleSlice returns a copy of rjsonMap with StdLibCompatibleString applied to all map keys and
// string values recursively.
func StdLibCompatibleMap(rjsonMap map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(rjsonMap))
	for key, val := range rjsonMap {
		k := StdLibCompatibleString(key)
		switch v := val.(type) {
		case string:
			out[k] = StdLibCompatibleString(v)
		case []interface{}:
			out[k] = StdLibCompatibleSlice(v)
		case map[string]interface{}:
			out[k] = StdLibCompatibleMap(v)
		default:
			out[k] = v
		}
	}
	return out
}
