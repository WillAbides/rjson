package rjson

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

func errUnexpectedByteInString(b byte) error {
	return fmt.Errorf("unexpected byte found in string: %q", string(b))
}

var (
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
//nolint:deadcode,unused,gocritic //copied code
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

func readFloat64Helper(hasDecimal, hasExp bool, digits []byte) (float64, error) {
	if hasDecimal || hasExp || len(digits) > 18 {
		return strconv.ParseFloat(unsafeBytesToString(digits), 64)
	}
	if digits[0] == '-' {
		n, err := readInt64Helper(true, digits[1:])
		return float64(n), err
	}
	n, err := readUint64Helper(digits)
	return float64(n), err
}

func unsafeBytesToString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b)) //nolint:gosec // ok
	stringHeader := reflect.StringHeader{Data: sliceHeader.Data, Len: sliceHeader.Len}
	return *(*string)(unsafe.Pointer(&stringHeader)) //nolint:gosec // ok
}

func readInt64Helper(neg bool, digits []byte) (int64, error) {
	const cutoff = uint64(1 << uint64(63))
	un, err := readUint64Helper(digits)
	if err != nil {
		return 0, err
	}
	if neg {
		if un > cutoff {
			return 0, fmt.Errorf(`value out of int64 range`)
		}
		return -int64(un), nil
	}
	if un >= cutoff {
		return 0, fmt.Errorf(`value out of int64 range`)
	}
	return int64(un), nil
}

func readUint64Helper(data []byte) (uint64, error) {
	val := uint64(0)
	const cutoff = (1<<64-1)/10 + 1
	const zero = uint64('0')
	ld := len(data)
	if ld > 19 {
		data = data[:19]
	}
	for _, digit := range data {
		v := uint64(digit) - zero
		val = (val * 10) + v
	}
	if ld > 19 {
		for _, digit := range data[19:ld] {
			v := uint64(digit) - zero
			if val > cutoff {
				return 0, fmt.Errorf(`value out of uint64 range`)
			}
			newVal := val * 10
			newVal += v
			if newVal < val {
				return 0, fmt.Errorf(`value out of uint64 range`)
			}
			val = newVal
		}
	}
	return val, nil
}
