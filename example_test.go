package rjson_test

import (
	"fmt"
	"log"

	"github.com/willabides/rjson"
)

func ExampleHandleObjectValues() {
	data := `
{
	"foo": "bar",
	"bar": [1,2,3],
    "baz": true
}
`
	var foo string
	var baz bool

	handler := rjson.ObjectValueHandlerFunc(func(fieldname, data []byte) (p int, err error) {
		switch string(fieldname) {
		case "foo":
			foo, p, err = rjson.ReadString(data, nil)
		case "baz":
			baz, p, err = rjson.ReadBool(data)
		}
		return p, err
	})
	var buffer rjson.Buffer
	p, err := rjson.HandleObjectValues([]byte(data), handler, &buffer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("We read %d bytes and got foo: %q, baz: %v\n", p, foo, baz)

	// Output: We read 52 bytes and got foo: "bar", baz: true
}
