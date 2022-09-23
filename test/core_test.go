package test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Bios-Marcel/yagcl"
	json "github.com/Bios-Marcel/yagcl-json"
	"github.com/stretchr/testify/assert"
)

func Test_JSONSource_InterfaceCompliance(t *testing.T) {
	var _ yagcl.Source = json.Source().Path("irrelevant.json")
}

func Test_JSONSource_ErrNoSource(t *testing.T) {
	source, ok := json.Source().(yagcl.Source)
	if assert.True(t, ok) {
		loaded, err := source.Parse(nil)
		assert.False(t, loaded)
		assert.ErrorIs(t, err, json.ErrNoDataSourceSpecified)
	}
}

func Test_JSONSource_MultipleSources(t *testing.T) {
	stepOne := json.Source()
	stepOne.Bytes([]byte{1})
	stepOne.Path("irrelevant.json")
	if source, ok := stepOne.(yagcl.Source); assert.True(t, ok) {
		loaded, err := source.Parse(nil)
		assert.False(t, loaded)
		assert.ErrorIs(t, err, json.ErrMultipleDataSourcesSpecified)
	}

	stepOne = json.Source()
	stepOne.String("{}")
	stepOne.Path("irrelevant.json")
	if source, ok := stepOne.(yagcl.Source); assert.True(t, ok) {
		loaded, err := source.Parse(nil)
		assert.False(t, loaded)
		assert.ErrorIs(t, err, json.ErrMultipleDataSourcesSpecified)
	}

	stepOne = json.Source()
	stepOne.Reader(bytes.NewReader([]byte{1}))
	stepOne.Path("irrelevant.json")
	if source, ok := stepOne.(yagcl.Source); assert.True(t, ok) {
		loaded, err := source.Parse(nil)
		assert.False(t, loaded)
		assert.ErrorIs(t, err, json.ErrMultipleDataSourcesSpecified)
	}

	stepOne = json.Source()
	stepOne.Reader(bytes.NewReader([]byte{1}))
	stepOne.Bytes([]byte{1})
	if source, ok := stepOne.(yagcl.Source); assert.True(t, ok) {
		loaded, err := source.Parse(nil)
		assert.False(t, loaded)
		assert.ErrorIs(t, err, json.ErrMultipleDataSourcesSpecified)
	}
}

func Test_Parse_StringSource(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().Add(json.
		Source().
		String(`{
			"field_a": "content a",
			"field_b": "content b"
		}`)).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content b", c.FieldB)
	}
}

func Test_Parse_PathSource(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().Add(json.Source().Path("./test.json")).Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content b", c.FieldB)
	}
}

func Test_Parse_PathSource_NotFound(t *testing.T) {
	type configuration struct{}
	var c configuration
	err := yagcl.New[configuration]().Add(json.Source().Path("./doesntexist.json").Must()).Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrSourceNotFound)
	err = yagcl.New[configuration]().Add(json.Source().Path("./doesntexist.json")).Parse(&c)
	assert.NoError(t, err)
}

func Test_Parse_PathSource_Dir(t *testing.T) {
	type configuration struct{}
	var c configuration
	err := yagcl.New[configuration]().Add(json.Source().Path("./").Must()).Parse(&c)
	assert.Error(t, err)
	err = yagcl.New[configuration]().Add(json.Source().Path("./")).Parse(&c)
	assert.Error(t, err)
}

func Test_Parse_ReaderSource(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB string `json:"field_b"`
	}
	var c configuration
	handle, errOpen := os.OpenFile("./test.json", os.O_RDONLY, os.ModePerm)
	if assert.NoError(t, errOpen) {
		err := yagcl.New[configuration]().Add(json.Source().Reader(handle)).Parse(&c)
		if assert.NoError(t, err) {
			assert.Equal(t, "content a", c.FieldA)
			assert.Equal(t, "content b", c.FieldB)
		}
	}
}

type failingReader struct {
	io.Reader
}

func (fr failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("you shall not ead")
}

func Test_Parse_ReaderSource_Error(t *testing.T) {
	type configuration struct{}
	var c configuration
	assert.Error(t, yagcl.New[configuration]().Add(json.Source().Reader(&failingReader{}).Must()).Parse(&c))
	assert.Error(t, yagcl.New[configuration]().Add(json.Source().Reader(&failingReader{})).Parse(&c))
}

func Test_Parse_KeyTags(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().
			Bytes([]byte(`{
				"field_a": "content a",
				"field_b": "content b"
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content b", c.FieldB)
	}
}

func Test_Parse_MissingJSONField(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().Add(json.Source().Bytes([]byte(`{}`))).Parse(&c)
	if assert.NoError(t, err) {
		assert.Empty(t, c.FieldA)
	}
}

func Test_Parse_MissingFieldKey(t *testing.T) {
	type configuration struct {
		FieldA string
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{"field_a": "content a"}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrExportedFieldMissingKey)
}

func Test_Parse_IgnoreField(t *testing.T) {
	type configuration struct {
		FieldA string `ignore:"true"`
		FieldB string `key:"field_b" ignore:"true"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().
			Bytes([]byte(`{
				"field_a": "content a",
				"field_b": "content b"
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Empty(t, c.FieldA)
	}
}

func Test_Parse_UnexportedFieldsIgnored(t *testing.T) {
	type configuration struct {
		fieldA string `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{"field_a": "content a"}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Empty(t, c.fieldA)
	}
}

func Test_Parse_TrailingCommas(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{
			"field_a": "content a",
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
	}
}

func Test_Parse_TrailingCommas_Array(t *testing.T) {
	// Sadly this test fails right now, it might be good to try and fix this.
	t.Skip()

	type configuration struct {
		FieldA []string `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{
			"field_a": ["content a",]
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, c.FieldA, []string{"content a"})
	}
}

func Test_Parse_TrailingCommas_Map(t *testing.T) {
	// Sadly this test fails right now, it might be good to try and fix this.
	t.Skip()

	type configuration struct {
		FieldA map[string]string `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{
			"field_a": {
				"a": "b",
			}
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, c.FieldA, map[string]string{"a": "b"})
	}
}

func Test_Parse_Comments(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB string `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{
			"field_a": "content a",
			//This is a comment
			"field_b": "content b"
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content b", c.FieldB)
	}
}

func Test_Parse_JSON5(t *testing.T) {
	// Simple JSON5 test. Right now this won't pass, but we'll keep it, so we
	// can check these at any time.
	t.SkipNow()

	type configuration struct {
		Unquoted            string   `key:"unquoted"`
		SingleQuotes        string   `key:"singleQuotes"`
		LineBreaks          string   `key:"lineBreaks"`
		Hexadecimal         string   `key:"hexadecimal"`
		LeadingDecimalPoint float64  `key:"leadingDecimalPoint"`
		AndTrailing         float64  `key:"andTrailing"`
		PositiveSign        int      `key:"positiveSign"`
		TrailingComma       string   `key:"trailingComma"`
		AndIn               []string `key:"andIn"`
		BackwardsCompatible string   `key:"backwardsCompatible"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Bytes([]byte(`{
			// comments
			unquoted: 'and you can quote me on that',
			singleQuotes: 'I can use "double quotes" here',
			lineBreaks: "Look, Mom! \
		  No \\n's!",
			hexadecimal: 0xdecaf,
			leadingDecimalPoint: .8675309, andTrailing: 8675309.,
			positiveSign: +1,
			trailingComma: 'in objects', andIn: ['arrays',],
			"backwardsCompatible": "with JSON",
		  }`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "and you can quote me on that", c.Unquoted)
		assert.Equal(t, `I can use "double quotes" here`, c.SingleQuotes)
		assert.Equal(t, "Look, Mom! \nNo \\n's!", c.LineBreaks)
		assert.Equal(t, 0xdecaf, c.Hexadecimal)
		assert.Equal(t, .8675309, c.LeadingDecimalPoint)
		assert.Equal(t, 8675309., c.AndTrailing)
		assert.Equal(t, 1, c.PositiveSign)
		assert.Equal(t, "in objects", c.TrailingComma)
		assert.Equal(t, []string{"arrays"}, c.AndIn)
		assert.Equal(t, "with JSON", c.BackwardsCompatible)
	}
}
