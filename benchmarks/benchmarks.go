package benchmarks

var benchers = []bencher{
	&jsonBencher{},
	&rjsonBencher{},
	&gjsonBencher{},
	&jsoniterBencher{},
	&jsonparserBencher{},
	&fastjsonBencher{},
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
