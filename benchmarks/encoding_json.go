package benchmarks

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type jsonBencher struct{}

func (x *jsonBencher) name() string {
	return "encoding_json"
}

func (x *jsonBencher) readFloat64(data []byte) (val float64, err error) {
	tkn, err := json.NewDecoder(bytes.NewReader(data)).Token()
	if err != nil {
		return 0, err
	}
	var ok bool
	val, ok = tkn.(float64)
	if ok {
		return val, nil
	}
	return 0, fmt.Errorf("not a number")
}

func (x *jsonBencher) readInt64(data []byte) (val int64, err error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	tkn, err := decoder.Token()
	if err != nil {
		return 0, err
	}
	num, ok := tkn.(json.Number)
	if ok {
		return num.Int64()
	}
	return 0, fmt.Errorf("not a number")
}

func (x *jsonBencher) decodeInt64(data []byte, v *int64) error {
	return json.Unmarshal(data, &v)
}

func (x *jsonBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (x *jsonBencher) valid(data []byte) bool {
	return json.Valid(data)
}

func (x *jsonBencher) readRepoData(data []byte, val *repoData) error {
	return json.Unmarshal(data, &val)
}

func (x *jsonBencher) readString(data []byte) (string, error) {
	tkn, err := json.NewDecoder(bytes.NewReader(data)).Token()
	if err != nil {
		return "", err
	}
	s, ok := tkn.(string)
	if ok {
		return s, nil
	}
	return "", fmt.Errorf("not a string")
}

func (x *jsonBencher) readBool(data []byte) (bool, error) {
	tkn, err := json.NewDecoder(bytes.NewReader(data)).Token()
	if err != nil {
		return false, err
	}
	b, ok := tkn.(bool)
	if ok {
		return b, nil
	}
	return false, fmt.Errorf("not a bool")
}
