package benchmarks

import (
	"github.com/valyala/fastjson"
)

type fastjsonBencher struct {
	parser *fastjson.Parser
}

func (x *fastjsonBencher) init() {
	*x = fastjsonBencher{
		parser: &fastjson.Parser{},
	}
}

func (x *fastjsonBencher) name() string {
	return "fastjson"
}

func (x *fastjsonBencher) readFloat64(data []byte) (val float64, err error) {
	parsed, err := x.parser.ParseBytes(data)
	if err != nil {
		return 0, err
	}
	return parsed.Float64()
}

func (x *fastjsonBencher) readInt64(data []byte) (val int64, err error) {
	parsed, err := x.parser.ParseBytes(data)
	if err != nil {
		return 0, err
	}
	return parsed.Int64()
}

func (x *fastjsonBencher) valid(data []byte) bool {
	err := fastjson.ValidateBytes(data)
	return err == nil
}

func (x *fastjsonBencher) readRepoData(data []byte, result *repoData) error {
	parser := x.parser
	val, err := parser.ParseBytes(data)
	if err != nil {
		return err
	}
	objectVal := val.Get("archived")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			result.Archived, err = objectVal.Bool()
			if err != nil {
				return err
			}
		}
	}

	objectVal = val.Get("full_name")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			sb, sbErr := objectVal.StringBytes()
			if sbErr != nil {
				return sbErr
			}
			result.FullName = string(sb)
		}
	}

	objectVal = val.Get("forks")
	if objectVal != nil {
		switch objectVal.Type() {
		case fastjson.TypeNull:
		default:
			result.Forks, err = objectVal.Int64()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (x *fastjsonBencher) readString(data []byte) (string, error) {
	parsed, err := x.parser.ParseBytes(data)
	if err != nil {
		return "", err
	}
	val, err := parsed.StringBytes()
	if err != nil {
		return "", err
	}
	return string(val), err
}

func (x *fastjsonBencher) readBool(data []byte) (bool, error) {
	parsed, err := x.parser.ParseBytes(data)
	if err != nil {
		return false, err
	}
	return parsed.Bool()
}
