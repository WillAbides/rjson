package benchmarks

import (
	"encoding/json"
)

type jsonBencher struct{}

func (x *jsonBencher) name() string {
	return "encoding_json"
}

func (x *jsonBencher) readFloat64(data []byte) (val float64, err error) {
	err = json.Unmarshal(data, &val)
	return val, err
}

func (x *jsonBencher) readInt64(data []byte) (val int64, err error) {
	err = json.Unmarshal(data, &val)
	return val, err
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
