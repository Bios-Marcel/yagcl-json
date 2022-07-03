package test

import (
	"bytes"
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
