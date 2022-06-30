package json

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

func Test_Parse_JSON_Simple(t *testing.T) {
	type configuration struct {
		//Not yet implemented
		//FieldA string `key:"field_a"`
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(json.Source().Path("./test.json").Must()).
		Parse(&c)
	assert.NoError(t, err)
	//Not yet implemented
	//assert.Equal(t, "content a", c.FieldA)
	assert.Equal(t, "content b", c.FieldB)
}
