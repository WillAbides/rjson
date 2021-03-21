package rjson

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ifaceUnmarshaller unmarshalls json to an interface
type ifaceUnmarshaller struct {
	val    interface{}
	buffer Buffer
}

func (h *ifaceUnmarshaller) HandleObjectValue(fieldname, data []byte) (int, error) {
	mpVal, ok := h.val.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("expected a map")
	}
	name, _, err := UnescapeStringContent(fieldname, nil)
	if err != nil {
		return 0, err
	}
	um := &ifaceUnmarshaller{}
	p, err := um.HandleAnyValue(data)
	if err != nil {
		return p, err
	}

	mpVal[string(name)] = um.val
	return p, nil
}

func (h *ifaceUnmarshaller) HandleValue(data []byte) (int, error) {
	sliceVal, ok := h.val.([]interface{})
	if !ok {
		return 0, fmt.Errorf("expected a slice")
	}
	um := &ifaceUnmarshaller{}
	p, err := um.HandleAnyValue(data)
	if err != nil {
		return p, err
	}
	sliceVal = append(sliceVal, um.val)
	h.val = sliceVal
	return p, err
}

func (h *ifaceUnmarshaller) HandleAnyValue(data []byte) (int, error) {
	tknType, p, err := NextTokenType(data)
	if err != nil {
		return p, err
	}
	p--
	data = data[p:]
	var pp int
	switch tknType {
	case InvalidType:
		err = fmt.Errorf("invalid type")
	case NullType:
		h.val = nil
		pp, err = ReadNull(data)
	case StringType:
		h.val, pp, err = ReadString(data, nil)
	case NumberType:
		h.val, pp, err = ReadFloat64(data)
	case TrueType, FalseType:
		h.val, pp, err = ReadBool(data)
	case ObjectStartType:
		h.val = map[string]interface{}{}
		pp, err = h.buffer.HandleObjectValues(data, h)
	case ArrayStartType:
		h.val = []interface{}{}
		pp, err = h.buffer.HandleArrayValues(data, h)
	default:
		err = fmt.Errorf("unexpected token: %s", tknType)
	}
	return p + pp, err
}

// removeJSONRuneError because encoding/json incorrectly unmarshals some characters to RuneError
// when we see that rjson differs from encoding/json change both values to ""  before comparing
func removeJSONRuneError(rjsonVal, jsonVal interface{}) (rvRes, jvRes interface{}) {
	var rvMap, jvMap map[string]interface{}
	var ok bool
	rvMap, ok = rjsonVal.(map[string]interface{})
	if ok {
		jvMap, ok = jsonVal.(map[string]interface{})
		if !ok {
			return rjsonVal, jsonVal
		}
		var jvv interface{}
		for k, rvv := range rvMap {
			jvv, ok = jvMap[k]
			if !ok {
				delete(rvMap, k)
				continue
			}
			rvMap[k], jvMap[k] = removeJSONRuneError(rvv, jvv)
		}
		for k := range jvMap {
			if !strings.ContainsRune(k, utf8.RuneError) {
				continue
			}
			_, ok = rvMap[k]
			if !ok {
				delete(jvMap, k)
			}
		}
		return rvMap, jvMap
	}

	var rvSlice, jvSlice []interface{}
	rvSlice, ok = rjsonVal.([]interface{})
	if ok {
		jvSlice, ok = jsonVal.([]interface{})
		if !ok {
			return rjsonVal, jsonVal
		}
		for i := range rvSlice {
			if i >= len(jvSlice) {
				break
			}
			rvSlice[i], jvSlice[i] = removeJSONRuneError(rvSlice[i], jvSlice[i])
		}
		return rvSlice, jvSlice
	}

	var rvStr, jvStr string
	rvStr, ok = rjsonVal.(string)
	if ok {
		jvStr, ok = jsonVal.(string)
		if !ok {
			return rjsonVal, jsonVal
		}
		if strings.ContainsRune(jvStr, utf8.RuneError) {
			jvStr = ""
			rvStr = ""
		}
		return jvStr, rvStr
	}

	return rjsonVal, jsonVal
}
