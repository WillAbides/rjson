package rjson

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	benchString string
	benchInt    int
	benchInt64  int64
	benchBool   bool
	benchBuf    Buffer
	benchFloat  float64
	benchFace   any
)

func BenchmarkSkip(b *testing.B) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))

		b.Run(file, func(b *testing.B) {
			b.Run("SkipValue", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(size)
				var err error
				for b.Loop() {
					benchInt, err = SkipValue(data, &benchBuf)
				}
				require.NoError(b, err)
			})

			b.Run("SkipValueFast", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(size)
				var err error
				for b.Loop() {
					benchInt, err = SkipValueFast(data, &benchBuf)
				}
				require.NoError(b, err)
			})
		})
	}
}

func BenchmarkValid(b *testing.B) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))
		b.Run(file, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(size)
			for b.Loop() {
				benchBool = Valid(data, &benchBuf)
			}
			require.True(b, benchBool)
		})
	}
}

func BenchmarkReadObject(b *testing.B) {
	for _, file := range jsonTestFiles {
		data := getTestdataJSONGz(b, file)
		size := int64(len(data))
		b.Run(file, func(b *testing.B) {
			var err error
			b.ReportAllocs()
			b.SetBytes(size)
			h := &ValueReader{}
			for b.Loop() {
				benchFace, _, err = h.ReadValue(data)
			}
			require.NoError(b, err)
		})
	}
}

func BenchmarkGetValuesFromObject(b *testing.B) {
	type resType struct {
		PublicGists int64  `json:"public_gists"`
		PublicRepos int64  `json:"public_repos"`
		Login       string `json:"login"`
	}

	wantRes := resType{
		PublicGists: 8,
		PublicRepos: 8,
		Login:       "octocat",
	}

	data := getTestdataJSONGz(b, "github_user.json")
	var res resType
	doneErr := fmt.Errorf("done")
	var err error
	buffer := &Buffer{}
	var stringBuf []byte
	var seenRepos, seenGists, seenLogin bool
	handler := ObjectValueHandlerFunc(func(fieldname, data []byte) (p int, err error) {
		switch string(fieldname) {
		case `public_gists`:
			res.PublicGists, p, err = ReadInt64(data)
			seenGists = true
		case `public_repos`:
			res.PublicRepos, p, err = ReadInt64(data)
			seenRepos = true
		case `login`:
			stringBuf, p, err = ReadStringBytes(data, stringBuf[:0])
			res.Login = string(stringBuf)
			seenLogin = true
		default:
			p, err = SkipValueFast(data, buffer)
		}
		if err == nil && seenGists && seenRepos && seenLogin {
			return p, doneErr
		}
		return p, err
	})
	b.ReportAllocs()
	for b.Loop() {
		seenGists, seenGists, seenLogin = false, false, false
		_, err = HandleObjectValues(data, handler, buffer)
	}
	require.Equal(b, wantRes, res)
	require.EqualError(b, err, "done")
}

func BenchmarkReadFloat64(b *testing.B) {
	datas := [][]byte{
		[]byte(`-123456789`),
		[]byte(`123`),
		[]byte(`123e45`),
		[]byte(`12.123456789012345`),
	}
	var err error
	b.ReportAllocs()
benchLoop:
	for b.Loop() {
		for _, data := range datas {
			benchFloat, _, err = ReadFloat64(data)
			if err != nil {
				break benchLoop
			}
		}
	}
	require.NoError(b, err)
	_ = benchFloat
}

func BenchmarkDecodeFloat64(b *testing.B) {
	datas := [][]byte{
		[]byte(`null`),
		[]byte(`-123456789`),
		[]byte(`123`),
		[]byte(`123e45`),
		[]byte(`12.123456789012345`),
	}
	var err error
	b.ReportAllocs()
benchLoop:
	for b.Loop() {
		for _, data := range datas {
			benchInt, err = DecodeFloat64(data, &benchFloat)
			if err != nil {
				break benchLoop
			}
		}
	}
	require.NoError(b, err)
	_ = benchFloat
}

func BenchmarkReadInt64(b *testing.B) {
	datas := [][]byte{
		[]byte(`-123456789`),
		[]byte(`123`),
		[]byte(`1234512345123451234`),
	}
	var err error
	b.ReportAllocs()
benchLoop:
	for b.Loop() {
		for _, data := range datas {
			benchInt64, _, err = ReadInt64(data)
			if err != nil {
				break benchLoop
			}
		}
	}
}

func BenchmarkDecodeInt64(b *testing.B) {
	datas := [][]byte{
		[]byte(`-123456789`),
		[]byte(`123`),
		[]byte(`1234512345123451234`),
	}
	var err error
	b.ReportAllocs()
benchLoop:
	for b.Loop() {
		for _, data := range datas {
			benchInt, err = DecodeInt64(data, &benchInt64)
			if err != nil {
				break benchLoop
			}
		}
	}
}

func BenchmarkReadString(b *testing.B) {
	simpleString := []byte(`"hello this is a string of somewhat normal length"`)
	complexString := []byte(
		`"@aym0566x \n\n名前:前田あゆみ\n第一印象:なんか怖っ！\n今の印象:とりあえずキモい。噛み合わない\n好きなところ:ぶすでキモいとこ😋✨✨\n思い出:んーーー、ありすぎ😊❤️\nLINE交換できる？:あぁ……ごめん✋\nトプ画をみて:照れますがな😘✨\n一言:お前は一生もんのダチ💖"`,
	)
	var err error

	b.Run("simple string", func(b *testing.B) {
		b.Run("nil buf", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(simpleString)))
			for b.Loop() {
				benchString, benchInt, err = ReadString(simpleString, nil)
			}
			require.NoError(b, err)
		})

		b.Run("with buf", func(b *testing.B) {
			var buf []byte
			b.ReportAllocs()
			b.SetBytes(int64(len(simpleString)))
			for b.Loop() {
				benchString, benchInt, err = ReadString(simpleString, &buf)
			}
			require.NoError(b, err)
		})
	})

	b.Run("complex string", func(b *testing.B) {
		b.Run("nil buf", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(complexString)))
			for b.Loop() {
				benchString, benchInt, err = ReadString(complexString, nil)
			}
			require.NoError(b, err)
		})

		b.Run("with buf", func(b *testing.B) {
			var buf []byte
			b.ReportAllocs()
			b.SetBytes(int64(len(complexString)))
			for b.Loop() {
				benchString, benchInt, err = ReadString(complexString, &buf)
			}
			require.NoError(b, err)
		})
	})
}

func BenchmarkDecodeString(b *testing.B) {
	data := []byte(
		`"@aym0566x \n\n名前:前田あゆみ\n第一印象:なんか怖っ！\n今の印象:とりあえずキモい。噛み合わない\n好きなところ:ぶすでキモいとこ😋✨✨\n思い出:んーーー、ありすぎ😊❤️\nLINE交換できる？:あぁ……ごめん✋\nトプ画をみて:照れますがな😘✨\n一言:お前は一生もんのダチ💖"`,
	)
	var err error
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	var stringBuf []byte
	for b.Loop() {
		benchInt, err = DecodeString(data, &benchString, &stringBuf)
	}
	require.NoError(b, err)
}

func BenchmarkReadStringBytes(b *testing.B) {
	data := []byte(
		`"@aym0566x \n\n名前:前田あゆみ\n第一印象:なんか怖っ！\n今の印象:とりあえずキモい。噛み合わない\n好きなところ:ぶすでキモいとこ😋✨✨\n思い出:んーーー、ありすぎ😊❤️\nLINE交換できる？:あぁ……ごめん✋\nトプ画をみて:照れますがな😘✨\n一言:お前は一生もんのダチ💖"`,
	)
	var buf []byte
	var err error
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for b.Loop() {
		buf, benchInt, err = ReadStringBytes(data, buf[:0])
	}
	require.NoError(b, err)
	_ = buf
}
