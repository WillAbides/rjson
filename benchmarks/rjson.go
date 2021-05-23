package benchmarks

import (
	"fmt"

	"github.com/willabides/rjson"
)

type rjsonBencher struct {
	valueReader            rjson.ValueReader
	buffer                 rjson.Buffer
	readRepoDataHandler    *rjsonReadRepoDataHandler
	stringBuffer           []byte
	distinctUserIDsHandler *rjsonDistinctUserIDsUserHandler
}

func (x *rjsonBencher) init() {
	*x = rjsonBencher{
		readRepoDataHandler: &rjsonReadRepoDataHandler{
			doneErr: fmt.Errorf("done"),
		},
		distinctUserIDsHandler: &rjsonDistinctUserIDsUserHandler{},
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

func (x *rjsonBencher) decodeInt64(data []byte, v *int64) error {
	_, err := rjson.DecodeInt64(data, v)
	return err
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

func (x *rjsonBencher) distinctUserIDs(data []byte, dest []int64) ([]int64, error) {
	x.distinctUserIDsHandler.userIDs = dest
	_, err := rjson.HandleObjectValues(data, x.distinctUserIDsHandler, &x.buffer)
	if err != nil {
		return nil, err
	}
	return x.distinctUserIDsHandler.userIDs, nil
}

type rjsonDistinctUserIDsUserHandler struct {
	statusesBuf rjson.Buffer
	arrBuf      rjson.Buffer
	userBuf     rjson.Buffer
	inUser      bool
	inStatus    bool
	userIDs     []int64
}

func (h *rjsonDistinctUserIDsUserHandler) HandleArrayValue(data []byte) (p int, err error) {
	h.inStatus = true
	p, err = rjson.HandleObjectValues(data, h, &h.arrBuf)
	h.inStatus = false
	return p, err
}

func (h *rjsonDistinctUserIDsUserHandler) HandleObjectValue(fieldname, data []byte) (p int, err error) {
	switch string(fieldname) {
	case "statuses":
		return rjson.HandleArrayValues(data, h, &h.statusesBuf)
	case "user":
		if h.inUser || !h.inStatus {
			return 0, nil
		}
		h.inUser = true
		p, err = rjson.HandleObjectValues(data, h, &h.userBuf)
		h.inUser = false
		return p, err
	case "id":
		if !h.inUser {
			return 0, nil
		}
		var v int64
		v, p, err = rjson.ReadInt64(data)
		h.userIDs = append(h.userIDs, v)
		return p, err
	}
	return 0, nil
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
