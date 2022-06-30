package json

import (
	"testing"

	"github.com/Bios-Marcel/yagcl"
	json "github.com/Bios-Marcel/yagcl-json"
	"github.com/stretchr/testify/assert"
)

func Test_JSONSource_InterfaceCompliance(t *testing.T) {
	var _ yagcl.Source = json.Source().Path("irrelevant.json")
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
