package benchmarks

import (
	"fmt"

	"github.com/willabides/rjson"
)

type rjsonBencher struct {
	valueReader  rjson.ValueReader
	buffer       rjson.Buffer
	stringBuffer []byte
	val          *rjson.JSONValue
}

func (x *rjsonBencher) init() {
	*x = rjsonBencher{
		val: &rjson.JSONValue{
			DoneErr:       fmt.Errorf("done"),
			RawFieldNames: true,
		},
	}
	x.val.AddObjectFieldValues(map[string]*rjson.JSONValue{
		"full_name": {},
		"archived":  {},
		"forks": {
			ParsedNumberType: rjson.JSONValueInt,
		},
	})
}

func (x *rjsonBencher) name() string {
	return "rjson"
}

func (x *rjsonBencher) readFloat64(data []byte) (val float64, err error) {
	val, _, err = rjson.ReadFloat64(data)
	return val, err
}

func (x *rjsonBencher) readInt64(data []byte) (val int64, err error) {
	val, _, err = rjson.ReadInt64(data)
	return val, err
}

func (x *rjsonBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	val, _, err = x.valueReader.ReadObject(data)
	return val, err
}

func (x *rjsonBencher) valid(data []byte) bool {
	return rjson.Valid(data, &x.buffer)
}

func (x *rjsonBencher) readRepoData(data []byte, val *repoData) (err error) {
	_, err = x.val.ParseJSON(data)
	if err == x.val.DoneErr {
		err = nil
	}
	if err != nil {
		return err
	}

	fields := x.val.Fields()
	fieldVal := fields["full_name"]
	if fieldVal.Exists() {
		switch fieldVal.TokenType() {
		case rjson.StringType:
			val.FullName = fieldVal.StringValue()
		case rjson.NullType:
		default:
			return fmt.Errorf("unexpected type")
		}
	}

	fieldVal = fields["archived"]
	if fieldVal.Exists() {
		switch fieldVal.TokenType() {
		case rjson.TrueType:
			val.Archived = true
		case rjson.FalseType:
			val.Archived = false
		case rjson.NullType:
		default:
			return fmt.Errorf("unexpected type")
		}
	}

	fieldVal = fields["forks"]
	if fieldVal.Exists() {
		switch fieldVal.TokenType() {
		case rjson.NumberType:
			val.Forks = fieldVal.IntValue()
		case rjson.NullType:
		default:
			return fmt.Errorf("unexpected type")
		}
	}

	return nil
}

func (x *rjsonBencher) readString(data []byte) (string, error) {
	var err error
	x.stringBuffer, _, err = rjson.ReadStringBytes(data, x.stringBuffer[:0])
	if err != nil {
		return "", err
	}
	return string(x.stringBuffer), nil
}

func (x *rjsonBencher) readBool(data []byte) (bool, error) {
	val, _, err := rjson.ReadBool(data)
	return val, err
}
