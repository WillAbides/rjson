package benchmarks

import (
	"fmt"

	"github.com/buger/jsonparser"
)

type jsonparserBencher struct {
	readRepoDataHandler *jsonparserReadRepoDataHandler
}

func (x *jsonparserBencher) init() {
	*x = jsonparserBencher{
		readRepoDataHandler: &jsonparserReadRepoDataHandler{
			doneErr: fmt.Errorf("done"),
		},
	}
}

func (x *jsonparserBencher) name() string {
	return "jsonparser"
}

func (x *jsonparserBencher) readFloat64(data []byte) (val float64, err error) {
	return jsonparser.GetFloat(data)
}

func (x *jsonparserBencher) readInt64(data []byte) (val int64, err error) {
	return jsonparser.GetInt(data)
}

func (x *jsonparserBencher) readRepoData(data []byte, result *repoData) error {
	h := x.readRepoDataHandler
	h.seenFullName, h.seenForks, h.seenArchived = false, false, false
	h.res = result
	err := jsonparser.ObjectEach(data, h.callback)
	if err == h.doneErr {
		err = nil
	}
	return err
}

type jsonparserReadRepoDataHandler struct {
	res          *repoData
	doneErr      error
	seenArchived bool
	seenForks    bool
	seenFullName bool
}

func (h *jsonparserReadRepoDataHandler) callback(key, value []byte, dataType jsonparser.ValueType, _ int) error {
	var err error
	switch string(key) {
	case "archived":
		h.seenArchived = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.Archived, err = jsonparser.ParseBoolean(value)
	case "forks":
		h.seenForks = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.Forks, err = jsonparser.ParseInt(value)
	case "full_name":
		h.seenFullName = true
		if dataType == jsonparser.Null {
			break
		}
		h.res.FullName, err = jsonparser.ParseString(value)
	}
	if err == nil && h.seenArchived && h.seenForks && h.seenFullName {
		return h.doneErr
	}
	return err
}

func (x *jsonparserBencher) readString(data []byte) (string, error) {
	return jsonparser.GetString(data)
}

func (x *jsonparserBencher) readBool(data []byte) (bool, error) {
	return jsonparser.GetBoolean(data)
}
