package benchmarks

import (
	"fmt"

	"github.com/minio/simdjson-go"
)

type simdjsonBencher struct {
	parsedJSON *simdjson.ParsedJson
	tmpIter    *simdjson.Iter
	obj        *simdjson.Object
	elem       *simdjson.Element
	array      *simdjson.Array
}

func (x *simdjsonBencher) name() string {
	return "simdjson"
}

func (x *simdjsonBencher) readRepoData(data []byte, val *repoData) error {
	var err error
	x.parsedJSON, err = simdjson.Parse(data, x.parsedJSON)
	if err != nil {
		return err
	}
	iter := x.parsedJSON.Iter()
	tp := iter.Advance()
	if tp != simdjson.TypeRoot {
		return fmt.Errorf("not root")
	}
	_, x.tmpIter, err = iter.Root(x.tmpIter)
	if err != nil {
		return err
	}
	x.obj, err = x.tmpIter.Object(x.obj)
	if err != nil {
		return err
	}

	elem := x.obj.FindKey("archived", x.elem)
	if elem != nil && elem.Type != simdjson.TypeNull {
		val.Archived, err = elem.Iter.Bool()
		if err != nil {
			return err
		}
		x.elem = elem
	}

	elem = x.obj.FindKey("forks", x.elem)
	if elem != nil && elem.Type != simdjson.TypeNull {
		val.Forks, err = elem.Iter.Int()
		if err != nil {
			return err
		}
		x.elem = elem
	}

	elem = x.obj.FindKey("full_name", x.elem)
	if elem != nil && elem.Type != simdjson.TypeNull {
		val.FullName, err = elem.Iter.String()
		if err != nil {
			return err
		}
		x.elem = elem
	}

	return nil
}

func (x *simdjsonBencher) distinctUserIDs(data []byte, dest []int64) ([]int64, error) {
	var err error
	x.parsedJSON, err = simdjson.Parse(data, x.parsedJSON)
	if err != nil {
		return nil, err
	}
	iter := x.parsedJSON.Iter()
	tp := iter.Advance()
	if tp != simdjson.TypeRoot {
		return nil, fmt.Errorf("not root")
	}
	_, x.tmpIter, err = iter.Root(x.tmpIter)
	if err != nil {
		return nil, err
	}
	x.obj, err = x.tmpIter.Object(x.obj)
	if err != nil {
		return nil, err
	}
	x.elem = x.obj.FindKey("statuses", x.elem)
	x.array, err = x.elem.Iter.Array(x.array)
	if err != nil {
		return nil, err
	}
	statusesIter := x.array.Iter()
	for statusesIter.Advance() == simdjson.TypeObject {
		x.obj, err = statusesIter.Object(x.obj)
		if err != nil {
			return nil, err
		}
		x.elem = x.obj.FindKey("user", nil)
		x.obj, err = x.elem.Iter.Object(x.obj)
		if err != nil {
			return nil, err
		}
		x.elem = x.obj.FindKey("id", x.elem)
		id, err := x.elem.Iter.Int()
		if err != nil {
			return nil, err
		}
		if id == 0 {
			continue
		}
		dest = append(dest, id)
	}
	return dest, nil
}

func (x *simdjsonBencher) readObject(data []byte) (val map[string]interface{}, err error) {
	x.parsedJSON, err = simdjson.Parse(data, x.parsedJSON)
	if err != nil {
		return nil, err
	}
	iter := x.parsedJSON.Iter()
	tp := iter.Advance()
	if tp != simdjson.TypeRoot {
		return nil, fmt.Errorf("not root")
	}
	_, x.tmpIter, err = iter.Root(x.tmpIter)
	if err != nil {
		return nil, err
	}
	x.obj, err = x.tmpIter.Object(x.obj)
	if err != nil {
		return nil, err
	}
	return x.obj.Map(nil)
}
