package rjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONValue_ParseJSON(t *testing.T) {
	data := getTestdataJSONGz(t, "github_repo.json")

	jv := &JSONValue{}
	jv.AddObjectFieldValues(map[string]*JSONValue{
		"archive_url":   {},
		"archived":      {},
		"forks":         {},
		"forks_count":   {ParsedNumberType: JSONValueRaw},
		"id":            {ParsedNumberType: JSONValueUint},
		"network_count": {ParsedNumberType: JSONValueInt},
	})
	jv.DoneErr = assert.AnError
	p, err := jv.ParseJSON(data)
	require.EqualError(t, err, assert.AnError.Error())
	require.Less(t, p, len(data))
	fields := jv.Fields()
	require.Equal(t, `https://api.github.com/repos/golang/go/{archive_format}{/ref}`, fields["archive_url"].StringValue())
	require.Equal(t, FalseType, fields["archived"].TokenType())
	require.Equal(t, float64(12162), fields["forks"].FloatValue())
	require.Equal(t, int64(0), fields["forks"].IntValue())
	require.Equal(t, uint64(0), fields["forks"].UintValue())
	require.Nil(t, fields["forks"].RawNumberValue())
	require.Equal(t, `12162`, string(fields["forks_count"].RawNumberValue()))
	require.Equal(t, float64(0), fields["forks_count"].FloatValue())
	require.Equal(t, uint64(23096959), fields["id"].UintValue())
	require.Equal(t, int64(12162), fields["network_count"].IntValue())
}
