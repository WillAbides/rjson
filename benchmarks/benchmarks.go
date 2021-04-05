package benchmarks

import (
	"os"
)

var benchers = []bencher{
	&jsonBencher{},
	&rjsonBencher{},
	&gjsonBencher{},
	&jsoniterBencher{},
	&jsonparserBencher{},
	&fastjsonBencher{},
}

func getBenchers(filter func(bencher) bool) []bencher {
	var bb []bencher
	bn := os.Getenv("BENCHER")
	for _, b := range benchers {
		if bn != "" && b.name() != bn {
			continue
		}
		if !filter(b) {
			continue
		}
		bb = append(bb, b)
	}
	return bb
}

func initBencher(b interface{}) {
	x, ok := b.(initter)
	if ok {
		x.init()
	}
}

type initter interface {
	init()
}

type bencher interface {
	name() string
}

type float64Reader interface {
	readFloat64(data []byte) (float64, error)
}

type int64Reader interface {
	readInt64(data []byte) (int64, error)
}

type objectReader interface {
	readObject(data []byte) (val map[string]interface{}, err error)
}

type validator interface {
	valid(data []byte) bool
}

type repoData struct {
	Archived bool   `json:"archived"`
	Forks    int64  `json:"forks"`
	FullName string `json:"full_name"`
}

type repoDataReader interface {
	readRepoData(data []byte, val *repoData) error
}

type stringReader interface {
	readString(data []byte) (string, error)
}

type boolReader interface {
	readBool(data []byte) (bool, error)
}
