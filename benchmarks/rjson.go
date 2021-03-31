package benchmarks

import (
	"fmt"

	"github.com/willabides/rjson"
)

type rjsonBencher struct {
	valueReader         rjson.ValueReader
	buffer              rjson.Buffer
	readRepoDataHandler *rjsonReadRepoDataHandler
	stringBuffer        []byte
}

func (x *rjsonBencher) init() {
	x.readRepoDataHandler = &rjsonReadRepoDataHandler{
		doneErr: fmt.Errorf("done"),
	}
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

func (x *rjsonBencher) readRepoData(data []byte, val *repoData) error {
	h := x.readRepoDataHandler
	h.seenFullName, h.seenForks, h.seenArchived = false, false, false
	h.res = val
	_, err := rjson.HandleObjectValues(data, h, &h.buffer)
	if err == h.doneErr {
		err = nil
	}
	return err
}

type rjsonReadRepoDataHandler struct {
	res          *repoData
	buffer       rjson.Buffer
	seenArchived bool
	seenForks    bool
	seenFullName bool
	doneErr      error
	stringBuf    []byte
}

func (h *rjsonReadRepoDataHandler) HandleObjectValue(fieldname, data []byte) (p int, err error) {
	var tknType rjson.TokenType
	tknType, _, err = rjson.NextTokenType(data)
	isNull := tknType == rjson.NullType
	if err != nil {
		return 0, err
	}
	switch string(fieldname) {
	case `archived`:
		h.seenArchived = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.res.Archived, p, err = rjson.ReadBool(data)
	case `forks`:
		h.seenForks = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.res.Forks, p, err = rjson.ReadInt64(data)
	case `full_name`:
		h.seenFullName = true
		if isNull {
			p, err = rjson.ReadNull(data)
			break
		}
		h.stringBuf, p, err = rjson.ReadStringBytes(data, h.stringBuf[:0])
		h.res.FullName = string(h.stringBuf)
	default:
		p, err = rjson.SkipValue(data, &h.buffer)
	}
	if err == nil && h.seenFullName && h.seenForks && h.seenArchived {
		return p, h.doneErr
	}
	return p, err
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
