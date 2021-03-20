package rjson

import (
	"fmt"
)

type TokenType uint8

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
