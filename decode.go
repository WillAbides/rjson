package rjson

import (
	"bytes"
	"encoding/json"
)

// DecodeBool reads a boolean value at the beginning of data. If data begins with null, v is untouched.
func DecodeBool(data []byte, v *bool) (p int, err error) {
	var val bool
	val, p, err = ReadBool(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeBoolCompat(data []byte, v *bool) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeFloat64 reads a float64 value at the beginning of data. If data begins with null, v is untouched.
func DecodeFloat64(data []byte, v *float64) (p int, err error) {
	var val float64
	val, p, err = ReadFloat64(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeFloat64Compat(data []byte, v *float64) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeInt64 reads an int64 value at the beginning of data. If data begins with null, v is untouched.
func DecodeInt64(data []byte, v *int64) (p int, err error) {
	var val int64
	val, p, err = ReadInt64(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeInt64Compat(data []byte, v *int64) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeInt32 reads an int32 value at the beginning of data. If data begins with null, v is untouched.
func DecodeInt32(data []byte, v *int32) (p int, err error) {
	var val int32
	val, p, err = ReadInt32(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeInt32Compat(data []byte, v *int32) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeInt reads an int value at the beginning of data. If data begins with null, v is untouched.
func DecodeInt(data []byte, v *int) (p int, err error) {
	var val int
	val, p, err = ReadInt(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeIntCompat(data []byte, v *int) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeUint64 reads a uint64 value at the beginning of data. If data begins with null, v is untouched.
func DecodeUint64(data []byte, v *uint64) (p int, err error) {
	var val uint64
	val, p, err = ReadUint64(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeUint64Compat(data []byte, v *uint64) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeUint32 reads a uint32 value at the beginning of data. If data begins with null, v is untouched.
func DecodeUint32(data []byte, v *uint32) (p int, err error) {
	var val uint32
	val, p, err = ReadUint32(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeUint32Compat(data []byte, v *uint32) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeUint reads a uint value at the beginning of data. If data begins with null, v is untouched.
func DecodeUint(data []byte, v *uint) (p int, err error) {
	var val uint
	val, p, err = ReadUint(data)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeUintCompat(data []byte, v *uint) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// DecodeString reads a string value at the beginning of data. If data begins with null, v is untouched.
// If buf is not nil, it will be used as a working for building the string value.
func DecodeString(data []byte, v *string, buf []byte) (p int, err error) {
	var val string
	val, p, err = ReadString(data, buf)
	if err != nil {
		return nullOrBust(data, err)
	}
	*v = val
	return p, err
}

func decodeStringCompat(data []byte, v *string) (p int, err error) {
	return decodeCompatHelper(data, v)
}

// nullOrBust tries to read "null" on data and returns origErr if it fails
func nullOrBust(data []byte, origErr error) (p int, err error) {
	p, err = ReadNull(data)
	if err != nil {
		return 0, origErr
	}
	return p, nil
}

func decodeCompatHelper(data []byte, v interface{}) (p int, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(v)
	if err != nil {
		return 0, err
	}
	return int(decoder.InputOffset()), nil
}
