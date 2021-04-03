package rjson

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextToken(t *testing.T) {
	t.Parallel()
	for _, td := range []struct {
		data string
		tkn  byte
		p    int
		err  string
	}{
		{data: `"foo"`, tkn: '"', p: 1},
		{data: `"`, tkn: '"', p: 1},
		{data: "\n\ntrue", tkn: 't', p: 3},
		{data: ``, err: "EOF"},
		{data: ` `, p: 1, err: "EOF"},
		{data: ` asdf `, tkn: 'a', p: 2, err: "no valid json token found"},
		{data: `asdf `, tkn: 'a', p: 1, err: "no valid json token found"},
	} {
		t.Run(td.data, func(t *testing.T) {
			tkn, p, err := NextToken([]byte(td.data))
			if td.err != "" {
				require.EqualError(t, err, td.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, string(td.tkn), string(tkn))
			require.Equal(t, td.p, p)
		})
	}
}

func TestNextTokenType(t *testing.T) {
	t.Parallel()
	for _, td := range []struct {
		data string
		tkn  TokenType
		p    int
		err  string
	}{
		{data: `"`, tkn: StringType, p: 1},
		{data: "true", tkn: TrueType, p: 1},
		{data: "false", tkn: FalseType, p: 1},
		{data: "null", tkn: NullType, p: 1},
		{data: "{", tkn: ObjectStartType, p: 1},
		{data: "}", tkn: ObjectEndType, p: 1},
		{data: "[", tkn: ArrayStartType, p: 1},
		{data: "]", tkn: ArrayEndType, p: 1},
		{data: ",", tkn: CommaType, p: 1},
		{data: ":", tkn: ColonType, p: 1},
		{data: "0", tkn: NumberType, p: 1},
		{data: ``, err: "EOF"},
		{data: ` `, p: 1, err: "EOF"},
		{data: ` asdf `, p: 2},
		{data: `asdf `, p: 1},
	} {
		t.Run(td.data, func(t *testing.T) {
			tkn, p, err := NextTokenType([]byte(td.data))
			if td.err != "" {
				require.EqualError(t, err, td.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, string(td.tkn), string(tkn))
			require.Equal(t, td.p, p)
		})
	}

	fileCounts := map[string]int{
		"canada.json":       334373,
		"citm_catalog.json": 135990,
		"twitter.json":      55263,
		"code.json":         396293,
		"example.json":      1297,
	}
	for filename, wantCount := range fileCounts {
		t.Run(filename, func(t *testing.T) {
			assert.Equal(t, wantCount, countTokens(getTestdataJSONGz(t, filename)))
		})
	}
}

func countTokens(data []byte) int {
	var count int
	buf := new(Buffer)
	for {
		tp, p, err := NextTokenType(data)
		if err != nil {
			break
		}
		count++
		data = data[p-1:]
		switch tp {
		case NullType, StringType, TrueType, FalseType, NumberType:
			p, err = SkipValueFast(data, buf)
			if err != nil {
				return count
			}
			data = data[p:]
		case InvalidType:
			fmt.Println(data[0])
			count--
			if len(data) > 0 {
				data = data[1:]
			}
		default:
			if len(data) > 0 {
				data = data[1:]
			}
		}
	}
	return count
}
