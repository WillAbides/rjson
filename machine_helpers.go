package rjson

import (
	"fmt"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func errUnexpectedByteInString(b byte) error {
	return fmt.Errorf("unexpected byte found in string: %q", string(b))
}

const skipMaxDepth = 10_000

var (
	errMaxDepth      = fmt.Errorf("exceeded max depth")
	errUnexpectedEOF = fmt.Errorf("unexpected end of json")
	errInvalidString = fmt.Errorf("invalid json string")
	errInvalidArray  = fmt.Errorf("invalid json array")
	errInvalidObject = fmt.Errorf("invalid json object")
	errInvalidUInt   = fmt.Errorf("invalid json uint")
	errInvalidInt    = fmt.Errorf("invalid json int")
	errInvalidNumber = fmt.Errorf("invalid json number")
	errNoValidToken  = fmt.Errorf("no valid json token found")
	errNotNull       = fmt.Errorf("not null")
	errNotBool       = fmt.Errorf("not a boolean value")
	errPOutOfRange   = fmt.Errorf("p out of range")
)

func growBytesSliceCapacity(slice []byte, size int) []byte {
	if cap(slice) >= size {
		return slice
	}
	origLen := len(slice)
	delta := size - cap(slice)
	slice = append(slice[:cap(slice)], make([]byte, delta)...)
	slice = slice[:origLen]

	return slice
}

func unescapeUnicodeChar(s, data []byte) (result []byte, bytesHandled int, ok bool) {
	rr := getu4(s)
	if rr < 0 {
		return data, 0, false
	}

	origLen := len(data)
	data = growBytesSliceCapacity(data, origLen+4)[:origLen+4]

	var w int
	if utf16.IsSurrogate(rr) {
		rr1 := getu4(s[6:])
		if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
			rl := utf8.RuneLen(dec)
			utf8.EncodeRune(data[origLen:], dec)
			return data[:origLen+rl], 12, true
		}
		rr = unicode.ReplacementChar
	}
	w = utf8.EncodeRune(data[origLen:], rr)
	return data[:origLen+w], 6, true
}

// copied from encoding/json
//
// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(data []byte) rune {
	if len(data) < 6 || data[0] != '\\' || data[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range data[2:6] {
		switch {
		case c >= '0' && c <= '9':
			c -= '0'
		case c >= 'a' && c <= 'f':
			c = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

// skipFloatExp skips the part of a float after the e
func skipFloatExp(data []byte) (p int, err error) {
	pe := len(data)
	if p == pe {
		return p, errInvalidNumber
	}
	sawDigit := false
	switch data[p] {
	case '+', '-':
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		sawDigit = true
	default:
		return p, errInvalidNumber
	}
	p++
	if p == pe {
		if sawDigit {
			return p, nil
		}
		return p, errInvalidNumber
	}
	switch data[p] {
	case '.', 'e', 'E':
		return p, errInvalidNumber
	case '0':
		if !sawDigit {
			return p, errInvalidNumber
		}
		sawDigit = true
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		sawDigit = true
	default:
		if sawDigit {
			return p, nil
		}
		return p, errInvalidNumber
	}
	p++
	for ; p < pe; p++ {
		switch data[p] {
		case '.', 'e', 'E':
			return p, errInvalidNumber
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return p, nil
		}
	}
	return p, nil
}

// skipFloatDec skips the part of a float after the decimal place
func skipFloatDec(data []byte) (p int, err error) {
	pe := len(data)
	if p == pe {
		return p, errInvalidNumber
	}
	switch data[p] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
	default:
		return p, errInvalidNumber
	}
	p++
	for ; p < pe; p++ {
		switch data[p] {
		case 'e', 'E':
			p++
			var pp int
			pp, err = skipFloatExp(data[p:])
			return p + pp, err
		case '.':
			return p, errInvalidNumber
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return p, nil
		}
	}
	return p, nil
}

func skipFloat(data []byte) (p int, err error) {
	pe := len(data)
	if p == pe {
		return p, errInvalidNumber
	}
	if data[p] == '-' {
		p++
		if p == pe {
			return p, errInvalidNumber
		}
	}
	switch data[p] {
	case '0':
		p++
		var pp int
		pp, err = skipFloatLeadingZero(data[p:])
		return p + pp, err
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
	default:
		return p, errInvalidNumber
	}
	p++
	for ; p < pe; p++ {
		switch data[p] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		case '.':
			p++
			var pp int
			pp, err = skipFloatDec(data[p:])
			return p + pp, err
		case 'e', 'E':
			p++
			var pp int
			pp, err = skipFloatLeadingZero(data[p:])
			return p + pp, err
		default:
			return p, nil
		}
	}
	return p, nil
}

func skipFloatLeadingZero(data []byte) (p int, err error) {
	pe := len(data)
	if p == pe {
		return p, nil
	}
	switch data[p] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'e', 'E':
		return p, errInvalidNumber
	case '.':
		var pp int
		p++
		pp, err = skipFloatDec(data[p:])
		return p + pp, err
	default:
		return p, nil
	}
}
