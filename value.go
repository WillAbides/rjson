package rjson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// JSONValueNumberType is the type of number parsed for a JSONValue
type JSONValueNumberType uint8

// JSONValueNumberType values
const (
	JSONValueFloat JSONValueNumberType = iota
	JSONValueInt
	JSONValueUint
	JSONValueRaw
)

// JSONValue is a parsed json value
type JSONValue struct {
	// DoneErr is the error HandleObjectValue should return after having seen all expected fields.
	// Set DoneErr to non-nil to make ParseJSON stop on done instead of reading the entire document
	// to the end. ParseJSON will return this error when returned by the handler.
	DoneErr error

	// ArrayValues are values from a json array. Make this the length of the array you expect to parse.
	// Array indexes with nil values in ArrayValues will not be parsed. See also AppendArrayValues.
	ArrayValues []*JSONValue

	// AppendArrayValues tells the handler to append ArrayValues when there are more values in an array
	// than values in ArrayValues. See also DefaultValue.
	AppendArrayValues bool

	// AddUnknownFields tells the handler to add any unknown field values it encounters when parsing an
	// object. See also DefaultValue.
	AddUnknownFields bool

	// DefaultValue is the value that is added to either Fields or ArrayValues when AddUnknownFields or
	// AppendArrayValues is true.
	DefaultValue *JSONValue

	// ParsedNumberType is the type of number value to parse. Default is JSONValueFloat
	ParsedNumberType JSONValueNumberType

	// RawFieldNames instructs the handler to not try to decode object field names
	RawFieldNames bool

	// RawStrings instructs the handler to not try decoding string values and return the raw string content instead.
	RawStrings bool

	// StdLibCompatibleObjectFields instructs the handler to use StdLibCompatibleString on field names.
	// This is ignored if RawFieldNames == true
	StdLibCompatibleObjectFields bool

	// StdLibCompatibleStrings instructs the handler to use StdLibCompatibleString on string values.
	// This is ignored if RawStrings == true
	StdLibCompatibleStrings bool

	fields          map[string]*JSONValue
	buf             Buffer
	foundFieldCount int
	avIndex         int

	found       bool
	depth       int
	tokenType   TokenType
	stringVal   []byte
	floatVal    float64
	parsedFloat bool
	intVal      int64
	parsedInt   bool
	uintVal     uint64
	parsedUint  bool
	rawNumber   []byte
	parsedRaw   bool
}

func (jv *JSONValue) parse(data []byte, depth int, doneErr error) (p int, err error) {
	jv.reset()
	jv.depth = depth
	jv.tokenType, p, err = NextTokenType(data)
	if err != nil {
		return p, err
	}
	jv.found = true
	p--
	var pp int
	switch jv.tokenType {
	case ObjectStartType:
		if jv.depth >= 10000 {
			return p, fmt.Errorf("exceeds max depth of 10000")
		}
		if jv.DoneErr == nil && doneErr != nil {
			jv.DoneErr = doneErr
			pp, err = HandleObjectValues(data[p:], jv, &jv.buf)
			jv.DoneErr = nil
			break
		}
		pp, err = HandleObjectValues(data[p:], jv, &jv.buf)
	case ArrayStartType:
		if jv.depth >= 10000 {
			return p, fmt.Errorf("exceeds max depth of 10000")
		}
		if jv.DoneErr == nil && doneErr != nil {
			jv.DoneErr = doneErr
			pp, err = HandleArrayValues(data[p:], jv, &jv.buf)
			jv.DoneErr = nil
			break
		}
		pp, err = HandleArrayValues(data[p:], jv, &jv.buf)
	case StringType:
		if jv.RawFieldNames {
			pp, err = SkipValue(data[p:], &jv.buf)
			jv.stringVal = data[p : p+pp]
			break
		}
		jv.stringVal, pp, err = ReadStringBytes(data[p:], jv.stringVal[:0])
		if jv.StdLibCompatibleStrings {
			jv.stringVal = StdLibCompatibleStringBytes(jv.stringVal, nil)
		}
	case NumberType:
		switch jv.ParsedNumberType {
		case JSONValueFloat:
			jv.floatVal, pp, err = ReadFloat64(data[p:])
			jv.parsedFloat = true
		case JSONValueInt:
			jv.intVal, pp, err = ReadInt64(data[p:])
			jv.parsedInt = true
		case JSONValueUint:
			jv.uintVal, pp, err = ReadUint64(data)
			jv.parsedUint = true
		case JSONValueRaw:
			pp, err = SkipValue(data[p:], &jv.buf)
			if err == nil {
				jv.rawNumber = data[p : p+pp]
			}
			jv.parsedRaw = true
		}
	case NullType, FalseType, TrueType:
		pp, err = SkipValue(data[p:], &jv.buf)
	default:
		pp, err = SkipValue(data[p:], &jv.buf)
	}
	return p + pp, err
}

// ParseJSON parses data and sets JSONValue accordingly
func (jv *JSONValue) ParseJSON(data []byte) (p int, err error) {
	return jv.parse(data, 0, nil)
}

// Fields are object fields.
func (jv *JSONValue) Fields() map[string]*JSONValue {
	return jv.fields
}

// AddObjectFieldValues adds object fields to be parsed.
func (jv *JSONValue) AddObjectFieldValues(fields map[string]*JSONValue) {
	if jv.fields == nil {
		jv.fields = make(map[string]*JSONValue, len(fields))
	}
	for k, v := range fields {
		jv.fields[k] = v
	}
}

// reset prepares *JSONValue to be parsed.
func (jv *JSONValue) reset() {
	for _, v := range jv.ArrayValues {
		v.reset()
	}
	for _, v := range jv.fields {
		v.reset()
	}
	jv.avIndex = 0
	jv.foundFieldCount = 0
	jv.found = false
	jv.rawNumber = jv.rawNumber[:0]
	jv.tokenType = InvalidType
	jv.parsedRaw, jv.parsedInt, jv.parsedFloat, jv.parsedUint = false, false, false, false
}

func (jv *JSONValue) prepFieldname(fieldname []byte) ([]byte, error) {
	var err error
	if bytes.IndexByte(fieldname, '\\') != -1 {
		fieldname, _, err = UnescapeStringContent(fieldname, fieldname[:0])
		if err != nil {
			return nil, err
		}
	}
	if !jv.StdLibCompatibleStrings {
		return fieldname, nil
	}
	return StdLibCompatibleStringBytes(fieldname, nil), nil
}

// HandleObjectValue implements ObjectValueHandler
func (jv *JSONValue) HandleObjectValue(fieldname, data []byte) (p int, err error) {
	if !jv.RawFieldNames {
		fieldname, err = jv.prepFieldname(fieldname)
		if err != nil {
			return 0, err
		}
	}
	v := jv.fields[string(fieldname)]
	if v == nil {
		if jv.fields == nil {
			jv.fields = make(map[string]*JSONValue)
		}
		if !jv.AddUnknownFields {
			return 0, nil
		}
		v = jv.getDefaultValue()
		jv.fields[string(fieldname)] = v
	}
	if !v.found {
		jv.foundFieldCount++
	}
	if v.found {
		v.reset()
	}
	var doneErr error
	if !jv.AddUnknownFields && jv.foundFieldCount == len(jv.fields) {
		doneErr = jv.DoneErr
	}
	p, err = v.parse(data, jv.depth+1, doneErr)
	if err == nil && !jv.AddUnknownFields && jv.foundFieldCount == len(jv.fields) {
		err = jv.DoneErr
	}
	return p, err
}

func (jv *JSONValue) clone() *JSONValue {
	if jv == nil {
		return nil
	}
	clone := *jv
	for i := range clone.ArrayValues {
		clone.ArrayValues[i] = clone.ArrayValues[i].clone()
	}
	for k := range clone.fields {
		clone.fields[k] = clone.fields[k].clone()
	}
	return &clone
}

func (jv *JSONValue) getDefaultValue() *JSONValue {
	if jv.DefaultValue != nil {
		return jv.DefaultValue.clone()
	}
	dv := JSONValue{
		AppendArrayValues:            jv.AppendArrayValues,
		AddUnknownFields:             jv.AddUnknownFields,
		StdLibCompatibleObjectFields: jv.StdLibCompatibleObjectFields,
		StdLibCompatibleStrings:      jv.StdLibCompatibleStrings,
		RawFieldNames:                jv.RawFieldNames,
		RawStrings:                   jv.RawStrings,
		ParsedNumberType:             jv.ParsedNumberType,
	}
	return &dv
}

func (jv *JSONValue) prepArray() (idx int, doneErr error) {
	avLen := len(jv.ArrayValues)
	if jv.avIndex < avLen {
		jv.avIndex++
		if jv.avIndex == avLen && !jv.AppendArrayValues {
			doneErr = jv.DoneErr
		}
		return jv.avIndex - 1, doneErr
	}
	if !jv.AppendArrayValues {
		return -1, nil
	}
	jv.ArrayValues = append(jv.ArrayValues, jv.getDefaultValue())
	jv.avIndex = avLen + 1
	return avLen, nil
}

// HandleArrayValue implements ArrayValueHandler
func (jv *JSONValue) HandleArrayValue(data []byte) (p int, err error) {
	idx, doneErr := jv.prepArray()
	if idx == -1 {
		return 0, jv.DoneErr
	}
	return jv.ArrayValues[idx].parse(data, jv.depth+1, doneErr)
}

// TokenType returns the TokenType associated with the parsed value.
func (jv *JSONValue) TokenType() TokenType {
	return jv.tokenType
}

// StringValueBytes returns the value of a parsed string as a byte slice.
func (jv *JSONValue) StringValueBytes() []byte {
	if !jv.found || jv.tokenType != StringType {
		return nil
	}
	return jv.stringVal
}

// StringValue returns the value of a parsed string
func (jv *JSONValue) StringValue() string {
	if !jv.found || jv.tokenType != StringType {
		return ""
	}
	return string(jv.stringVal)
}

// RawNumberValue returns the raw number value found when ParsedNumberType == JSONValueRaw
func (jv *JSONValue) RawNumberValue() []byte {
	if !jv.found || !jv.parsedRaw {
		return nil
	}
	return jv.rawNumber
}

// IntValue returns the int64 parsed when ParsedNumberType == JSONValueInt
func (jv *JSONValue) IntValue() int64 {
	if !jv.found || !jv.parsedInt {
		return 0
	}
	return jv.intVal
}

// UintValue returns the uint64 parsed when ParsedNumberType == JSONValueUint
func (jv *JSONValue) UintValue() uint64 {
	if !jv.found || !jv.parsedUint {
		return 0
	}
	return jv.uintVal
}

// FloatValue returns the float64 parsed when ParsedNumberType == JSONValueFloat
func (jv *JSONValue) FloatValue() float64 {
	if !jv.found || !jv.parsedFloat {
		return 0
	}
	return jv.floatVal
}

// Exists returns true if the value was found when parsing.
func (jv *JSONValue) Exists() bool {
	if jv == nil {
		return false
	}
	return jv.found
}

// This only exists for round trip fuzzing. Nothing else should call it.
func (jv *JSONValue) toInterface() interface{} {
	if !jv.found {
		return nil
	}
	switch jv.tokenType {
	case StringType:
		return jv.StringValue()
	case NumberType:
		switch {
		case jv.parsedRaw:
			return json.Number(jv.RawNumberValue())
		case jv.parsedInt:
			return float64(jv.IntValue())
		case jv.parsedUint:
			return float64(jv.UintValue())
		case jv.parsedFloat:
			return jv.FloatValue()
		}
	case NullType:
		return nil
	case TrueType:
		return true
	case FalseType:
		return false
	case ObjectStartType:
		mp := make(map[string]interface{}, len(jv.fields))
		for k, value := range jv.fields {
			if !value.found {
				continue
			}
			mp[k] = value.toInterface()
		}
		return mp
	case ArrayStartType:
		sl := make([]interface{}, len(jv.ArrayValues))
		for i := 0; i < len(jv.ArrayValues); i++ {
			sl[i] = jv.ArrayValues[i].toInterface()
		}
		return sl
	}
	return nil
}
