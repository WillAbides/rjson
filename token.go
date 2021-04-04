package rjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// TokenType is the type of a json token
type TokenType uint8

// TokenTypes
const (
	InvalidType TokenType = iota
	NullType
	StringType
	NumberType
	TrueType
	FalseType
	ObjectStartType
	ObjectEndType
	ArrayStartType
	ArrayEndType
	CommaType
	ColonType
)

var tokenTypeStrings = [256]string{
	InvalidType:     "invalid",
	NullType:        "null",
	StringType:      "string",
	NumberType:      "number",
	TrueType:        "true",
	FalseType:       "false",
	ObjectStartType: "object start",
	ObjectEndType:   "object end",
	ArrayStartType:  "array start",
	ArrayEndType:    "array end",
	CommaType:       "comma",
	ColonType:       "colon",
}

func (t TokenType) String() string {
	s := tokenTypeStrings[t]
	if s == "" {
		s = fmt.Sprintf("unknown type (%d)", t)
	}
	return s
}

var tokenTypes = [256]TokenType{
	'n': NullType,
	'"': StringType,
	't': TrueType,
	'f': FalseType,
	'{': ObjectStartType,
	'}': ObjectEndType,
	'[': ArrayStartType,
	']': ArrayEndType,
	'-': NumberType,
	'0': NumberType,
	'1': NumberType,
	'2': NumberType,
	'3': NumberType,
	'4': NumberType,
	'5': NumberType,
	'6': NumberType,
	'7': NumberType,
	'8': NumberType,
	'9': NumberType,
	',': CommaType,
	':': ColonType,
}

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

func nextTokenCompat(data []byte) (token byte, p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	tkn, err := decoder.Token()
	p = int(decoder.InputOffset())
	if err != nil {
		if p >= len(data) {
			return 0, p, err
		}
		// decoder.Token() validates, so check for tokens even if it errors
		for _, b := range `}],:"tfn-0123456789` {
			if data[p] == byte(b) {
				tkn = data[p]
				p++
				err = nil
				break
			}
		}
		switch err {
		case nil:
		case io.EOF:
			return 0, p, err
		default:
			return data[p], p, err
		}
	}
	switch tt := tkn.(type) {
	case byte:
		return tt, p, nil
	case json.Delim:
		return byte(tt), p, nil
	case bool:
		if tt {
			return 't', p - len(`true`) + 1, nil
		}
		return 'f', p - len(`false`) + 1, nil
	case json.Number:
		return tt.String()[0], p - len(tt.String()) + 1, nil
	case string:
		return '"', bytes.IndexByte(data, '"') + 1, nil
	case nil:
		return 'n', p - len(`null`) + 1, nil
	}
	return 0, p, errNoValidToken
}

// NextTokenType finds the first json token in data and returns its TokenType. p is the position in data immediately
// after where the token was found. NextToken errors if it finds anything besides json whitespace before the first valid
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

func nextTokenTypeCompat(data []byte) (TokenType, int, error) {
	var p int
	var err error
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	tkn, err := decoder.Token()
	p = int(decoder.InputOffset())
	if err != nil {
		if err == io.EOF {
			return 0, 0, io.EOF
		}
		if p >= len(data) {
			return InvalidType, p, nil
		}
		// decoder.Token() validates, so check for tokens even if it errors
		for _, b := range `}],:"tfn-0123456789` {
			if data[p] == byte(b) {
				tkn = data[p]
				p++
				err = nil
				break
			}
		}
		if err != nil {
			return InvalidType, p + 1, nil
		}
	}

	switch tt := tkn.(type) {
	case byte:
		return tokenTypes[tt], p, nil
	case json.Delim:
		return tokenTypes[byte(tt)], p, nil
	case bool:
		if tt {
			return TrueType, p - len(`true`) + 1, nil
		}
		return FalseType, p - len(`false`) + 1, nil
	case json.Number:
		return NumberType, p - len(tt.String()) + 1, nil
	case string:
		return StringType, bytes.IndexByte(data, '"') + 1, nil
	case nil:
		return NullType, p - len(`null`) + 1, nil
	}
	return InvalidType, p, nil
}
