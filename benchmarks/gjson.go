package benchmarks

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type gjsonBencher struct {
	idMap map[int64]struct{}
}

func (x *gjsonBencher) init() {
	*x = gjsonBencher{
		idMap: map[int64]struct{}{},
	}
}

func (x *gjsonBencher) name() string {
	return "gjson"
}

func (x *gjsonBencher) readFloat64(data []byte) (val float64, err error) {
	result := gjson.ParseBytes(data)
	if result.Type != gjson.Number {
		return 0, fmt.Errorf("not a number")
	}
	return result.Num, nil
}

func (x *gjsonBencher) readInt64(data []byte) (val int64, err error) {
	result := gjson.ParseBytes(data)
	switch result.Type {
	case gjson.Null:
		return 0, fmt.Errorf("null is not a number")
	case gjson.Number:
		val := result.Int()
		if float64(val) == result.Num {
			// risks overflowing on 32 bit systems, but close enough for now
			return val, nil
		}
		return 0, fmt.Errorf("not an int")
	}
	return 0, fmt.Errorf("not an int")
}

func (x *gjsonBencher) decodeInt64(data []byte, v *int64) error {
	result := gjson.ParseBytes(data)
	switch result.Type {
	case gjson.Null:
		return nil
	case gjson.Number:
		val := result.Int()
		if float64(val) == result.Num {
			*v = val
			return nil
		}
	}
	return fmt.Errorf("not an int")
}

func (x *gjsonBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	result := gjson.ParseBytes(data)
	mp, ok := result.Value().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("not a map")
	}
	return mp, nil
}

func (x *gjsonBencher) valid(data []byte) bool {
	return gjson.ValidBytes(data)
}

func (x *gjsonBencher) readRepoData(data []byte, result *repoData) error {
	results := gjson.GetManyBytes(data, `archived`, `full_name`, `forks`)

	switch results[0].Type {
	case gjson.True, gjson.False:
		result.Archived = results[0].Bool()
	case gjson.Null:
	default:
		return fmt.Errorf("unexpected type")
	}

	switch results[1].Type {
	case gjson.String:
		result.FullName = results[1].Str
	case gjson.Null:
	default:
		return fmt.Errorf("unexpected type")
	}

	switch results[2].Type {
	case gjson.Number:
		result.Forks = results[2].Int()
	case gjson.Null:
	default:
		return fmt.Errorf("unexpected type")
	}
	return nil
}

func (x *gjsonBencher) readString(data []byte) (string, error) {
	result := gjson.ParseBytes(data)
	val, ok := result.Value().(string)
	if !ok {
		return "", fmt.Errorf("not a string")
	}
	if !gjson.ValidBytes(data) {
		return "", fmt.Errorf("not a valid string")
	}
	return val, nil
}

func (x *gjsonBencher) readBool(data []byte) (bool, error) {
	if !gjson.ValidBytes(data) {
		return false, fmt.Errorf("not a bool")
	}
	result := gjson.ParseBytes(data)
	val, ok := result.Value().(bool)
	if !ok {
		return false, fmt.Errorf("not a bool")
	}
	return val, nil
}

func (x *gjsonBencher) distinctUserIDs(data []byte, dest []int64) ([]int64, error) {
	result := gjson.GetBytes(data, "statuses.#.user.id")
	for _, r := range result.Array() {
		dest = append(dest, r.Int())
	}
	return dest, nil
}
