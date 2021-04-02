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

func skipFloatExp(data []byte, p, pe int) (int, error) {
	if p == pe {
		return p - 1, errInvalidNumber
	}
	startP := p
	signed := false
	for ; p < pe; p++ {
		if digits[data[p]] {
			continue
		}
		if p == startP && signBytes[data[p]] {
			signed = true
			continue
		}
		break
	}
	var err error
	switch p - startP {
	case 0:
		err = errInvalidNumber
	case 1:
		if signed {
			err = errInvalidNumber
		}
	}
	return p - 1, err
}

func skipFloatDec(data []byte, p, pe int) (int, error) {
	if p == pe {
		return p - 1, errInvalidNumber
	}
	if !digits[data[p]] {
		return p - 1, errInvalidNumber
	}
	p++
	for ; p < pe; p++ {
		if !digits[data[p]] {
			break
		}
	}
	if p == pe {
		return p - 1, nil
	}

	if expBytes[data[p]] {
		return skipFloatExp(data, p+1, pe)
	}

	return p - 1, nil
}

var signBytes = [256]bool{
	'-': true,
	'+': true,
}

var expBytes = [256]bool{
	'e': true,
	'E': true,
}

var digits = [256]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
}
