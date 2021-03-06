package benchmarks

import (
	goccyjson "github.com/goccy/go-json"
)

type goccyjsonBencher struct{}

func (x *goccyjsonBencher) name() string {
	return "goccyjson"
}

func (x *goccyjsonBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	err = goccyjson.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (x *goccyjsonBencher) readRepoData(data []byte, val *repoData) error {
	return goccyjson.Unmarshal(data, &val)
}
