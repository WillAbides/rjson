package rjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONValue_ParseJSON(t *testing.T) {
	data := getTestdataJSONGz(t, "github_repo.json")

	jv := &JSONValue{}
	var archiveURL, archived, forks JSONValue
	forksCount := JSONValue{ParsedNumberType: JSONValueRaw}
	id := JSONValue{ParsedNumberType: JSONValueUint}
	networkCount := JSONValue{ParsedNumberType: JSONValueInt}
	jv.AddObjectFieldValues(map[string]JSONParser{
		"archive_url":   &archiveURL,
		"archived":      &archived,
		"forks":         &forks,
		"forks_count":   &forksCount,
		"id":            &id,
		"network_count": &networkCount,
	})
	jv.DoneErr = assert.AnError
	p, err := jv.ParseJSON(data, nil)
	require.EqualError(t, err, assert.AnError.Error())
	require.Less(t, p, len(data))
	require.Equal(t, `https://api.github.com/repos/golang/go/{archive_format}{/ref}`, archiveURL.StringValue())
	require.Equal(t, FalseType, archived.TokenType())
	require.Equal(t, float64(12162), forks.FloatValue())
	require.Equal(t, int64(0), forks.IntValue())
	require.Equal(t, uint64(0), forks.UintValue())
	require.Nil(t, forks.RawNumberValue())
	require.Equal(t, `12162`, string(forksCount.RawNumberValue()))
	require.Equal(t, float64(0), forksCount.FloatValue())
	require.Equal(t, uint64(23096959), id.UintValue())
	require.Equal(t, int64(12162), networkCount.IntValue())
}
