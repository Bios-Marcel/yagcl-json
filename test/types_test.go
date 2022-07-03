package test

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Bios-Marcel/yagcl"
	yagcl_json "github.com/Bios-Marcel/yagcl-json"
	"github.com/stretchr/testify/assert"
)

func Test_Parse_JSON_String(t *testing.T) {
	type configuration struct {
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.Source().Bytes([]byte(`{"field_b": "content b"}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content b", c.FieldB)
	}
}

func Test_Parse_JSON_String_Invalid(t *testing.T) {
	type configuration struct {
		FieldB string `json:"field_b"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.Source().Bytes([]byte(`{"field_b": text}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Duration(t *testing.T) {
	type configuration struct {
		FieldA time.Duration `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.Source().Bytes([]byte(`{"field_a": "10s"}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, time.Second*10, c.FieldA)
	}
}

func Test_Parse_Duration_Invalid(t *testing.T) {
	type configuration struct {
		FieldA time.Duration `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.Source().Bytes([]byte(`{"field_a": "ain't valid"}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_JSON_Nested(t *testing.T) {
	type configuration struct {
		//Not yet implemented
		//FieldA string `key:"field_a"`
		FieldA struct {
			FieldB string `json:"field_b"`
		} `json:"field_a"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.Source().
			Bytes([]byte(`{
				"field_a": {
					"field_b": "content b"
				}
			}`)).Must()).
		Parse(&c)
	if assert.NoError(t, err) {
		//Not yet implemented
		//assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content b", c.FieldA.FieldB)
	}
}

func Test_Parse_DeeplyNested(t *testing.T) {
	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB struct {
			FieldC struct {
				FieldD struct {
					FieldE struct {
						FieldF struct {
							FieldG string `key:"field_g"`
						} `key:"field_f"`
					} `key:"field_e"`
				} `key:"field_d"`
			} `key:"field_c"`
		} `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().Add(yagcl_json.Source().
		Bytes([]byte(`
			{
				"field_a": "content a",
				"field_b": {
					"field_c": {
						"field_d": {
							"field_e": {
								"field_f": {
									"field_g": "content g"
								}
							}
						}
					}
				}
			}
		`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content g", c.FieldB.FieldC.FieldD.FieldE.FieldF.FieldG)
	}
}

func Test_Parse_SimplePointer(t *testing.T) {
	type configuration struct {
		FieldA *uint `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": 10
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint(10), *c.FieldA)
	}
}

func Test_Parse_DoublePointer(t *testing.T) {
	type configuration struct {
		FieldA **uint `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": 10
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint(10), **c.FieldA)
	}
}

func Test_Parse_PointerOfDoom(t *testing.T) {
	type configuration struct {
		FieldA ***************************************uint `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": 10
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint(10), ***************************************c.FieldA)
	}
}

func Test_Parse_SinglePointerToStruct(t *testing.T) {
	type substruct struct {
		FieldC string `key:"field_c"`
	}
	type configuration struct {
		FieldA string     `key:"field_a"`
		FieldB *substruct `key:"field_b"`
	}

	var c configuration
	c.FieldB = &substruct{}
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": "content a",
				"field_b": {
					"field_c": "content c"
				}
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content c", (*c.FieldB).FieldC)
	}
}

func Test_Parse_SinglePointerToStruct_Invalid(t *testing.T) {
	//FIXME
	t.SkipNow()

	type substruct struct {
		FieldC int `key:"field_c"`
	}
	type configuration struct {
		FieldA string     `key:"field_a"`
		FieldB *substruct `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
			"field_a": "content a",
			"field_b": {
				"field_c": "no integer here"
			}
		}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Struct_Invalid(t *testing.T) {
	type substruct struct {
		FieldC int `key:"field_c"`
	}
	type configuration struct {
		FieldA string    `key:"field_a"`
		FieldB substruct `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
			"field_a": "content a",
			"field_b": {
				"field_c": "no integer here"
			}
		}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_SingleNilPointerToStruct(t *testing.T) {
	//FIXME
	t.SkipNow()

	type substruct struct {
		FieldC string `key:"field_c"`
	}
	type configuration struct {
		FieldA string     `key:"field_a"`
		FieldB *substruct `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
			"field_a": "content a",
			"field_b": {
				"field_c": "content c"
			}
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content c", (*c.FieldB).FieldC)
	}
}

func Test_Parse_PointerOfDoomToStruct(t *testing.T) {
	//FIXME
	t.SkipNow()

	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB **************struct {
			FieldC string `key:"field_c"`
		} `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
			"field_a": "content a",
			"field_b": {
				"field_c": "content c"
			}
		}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		assert.Equal(t, "content c", (**************c.FieldB).FieldC)
	}
}

func Test_Parse_NestedPointerOfDoomToStruct(t *testing.T) {
	//FIXME
	t.SkipNow()

	type configuration struct {
		FieldA string `key:"field_a"`
		FieldB **************struct {
			FieldC **************struct {
				FieldD **************struct {
					FieldE string `key:"field_e"`
				} `key:"field_d"`
			} `key:"field_c"`
		} `key:"field_b"`
	}

	var c configuration
	err := yagcl.New[configuration]().Add(yagcl_json.Source().
		Bytes([]byte(`
			{
				"field_a": "content a",
				"field_b": {
					"field_c": {
						"field_d": {
							"field_e": "content e"
						}
					}
				}
			}
		`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, "content a", c.FieldA)
		fieldC := (**************c.FieldB).FieldC
		fieldD := (**************fieldC).FieldD
		fieldE := (**************fieldD).FieldE
		assert.Equal(t, "content e", fieldE)
	}
}

func Test_Parse_TypeAlias_NoCustomUnmarshal(t *testing.T) {
	type noopstring string
	type configuration struct {
		FieldA noopstring `key:"field_a"`
	}
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": "lower"
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, noopstring("lower"), c.FieldA)
	}
}

type uppercaser string

func (uc *uppercaser) UnmarshalJSON(data []byte) error {
	var intermediate string
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}

	if intermediate != "" {
		*uc = uppercaser(strings.ToUpper(intermediate))
	}

	return nil
}

func Test_Parse_CustomUnmarshaler(t *testing.T) {
	type customUnmarshalTest struct {
		FieldA uppercaser `json:"field_a"`
	}

	var customUnmarshalTestValue customUnmarshalTest
	if assert.NoError(t,
		json.Unmarshal(
			[]byte(`{
				"field_a": "lower"
			}`),
			&customUnmarshalTestValue)) {
		assert.Equal(t, uppercaser("LOWER"), customUnmarshalTestValue.FieldA)
	}

	type configuration struct {
		FieldA uppercaser `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": "lower"
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uppercaser("LOWER"), c.FieldA)
	}
}

func Test_Parse_Complex64_Unsupported(t *testing.T) {
	type configuration struct {
		FieldA complex64 `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": "irrelevant"
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrUnsupportedFieldType)
}

func Test_Parse_Complex128_Unsupported(t *testing.T) {
	type configuration struct {
		FieldA complex128 `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": "irrelevant"
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrUnsupportedFieldType)
}

func Test_Parse_Bool_Valid(t *testing.T) {
	type configuration struct {
		A bool `key:"a"`
		B bool `key:"b"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"a": true,
				"b": false
			}`))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, true, c.A)
		assert.Equal(t, false, c.B)
	}
}

func Test_Parse_Bool_Invalid(t *testing.T) {
	type configuration struct {
		Bool bool `key:"bool"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"bool": "cheese"
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Int_Valid(t *testing.T) {
	type configuration struct {
		Min int `key:"min"`
		Max int `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": %d,
				"max": %d
			}`, math.MinInt, math.MaxInt)))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, math.MinInt, c.Min)
		assert.Equal(t, math.MaxInt, c.Max)
	}
}

func Test_Parse_Int8_Valid(t *testing.T) {
	type configuration struct {
		Min int8 `key:"min"`
		Max int8 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": %d,
				"max": %d
			}`, math.MinInt8, math.MaxInt8)))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, int8(math.MinInt8), c.Min)
		assert.Equal(t, int8(math.MaxInt8), c.Max)
	}
}

func Test_Parse_Int16_Valid(t *testing.T) {
	type configuration struct {
		Min int16 `key:"min"`
		Max int16 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": %d,
				"max": %d
			}`, math.MinInt16, math.MaxInt16)))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, int16(math.MinInt16), c.Min)
		assert.Equal(t, int16(math.MaxInt16), c.Max)
	}
}

func Test_Parse_Int32_Valid(t *testing.T) {
	type configuration struct {
		Min int32 `key:"min"`
		Max int32 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": %d,
				"max": %d
			}`, math.MinInt32, math.MaxInt32)))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, int32(math.MinInt32), c.Min)
		assert.Equal(t, int32(math.MaxInt32), c.Max)
	}
}

func Test_Parse_Int64_Valid(t *testing.T) {
	type configuration struct {
		Min int64 `key:"min"`
		Max int64 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": %d,
				"max": %d
			}`, math.MinInt64, math.MaxInt64)))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(math.MinInt64), c.Min)
		assert.Equal(t, int64(math.MaxInt64), c.Max)
	}
}

func Test_Parse_Uint_Valid(t *testing.T) {
	type configuration struct {
		Min uint `key:"min"`
		Max uint `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": 0,
				"max": %d
			}`, uint64(math.MaxUint))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint(0), c.Min)
		assert.Equal(t, uint(math.MaxUint), c.Max)
	}
}

func Test_Parse_Uint8_Valid(t *testing.T) {
	type configuration struct {
		Min uint8 `key:"min"`
		Max uint8 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": 0,
				"max": %d
			}`, uint64(math.MaxUint8))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint8(0), c.Min)
		assert.Equal(t, uint8(math.MaxUint8), c.Max)
	}
}

func Test_Parse_Uint16_Valid(t *testing.T) {
	type configuration struct {
		Min uint16 `key:"min"`
		Max uint16 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": 0,
				"max": %d
			}`, uint64(math.MaxUint16))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(0), c.Min)
		assert.Equal(t, uint16(math.MaxUint16), c.Max)
	}
}

func Test_Parse_Uint32_Valid(t *testing.T) {
	type configuration struct {
		Min uint32 `key:"min"`
		Max uint32 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": 0,
				"max": %d
			}`, uint64(math.MaxUint32))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint32(0), c.Min)
		assert.Equal(t, uint32(math.MaxUint32), c.Max)
	}
}

func Test_Parse_Uint64_Valid(t *testing.T) {
	type configuration struct {
		Min uint64 `key:"min"`
		Max uint64 `key:"max"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"min": 0,
				"max": %d
			}`, uint64(math.MaxUint64))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, uint64(0), c.Min)
		assert.Equal(t, uint64(math.MaxUint64), c.Max)
	}
}

func Test_Parse_Float32_Valid(t *testing.T) {
	type configuration struct {
		Float float32 `key:"float"`
	}

	var floatValue float32 = 5.5
	bytes, _ := json.Marshal(floatValue)
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"float": %s
			}`, string(bytes))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, floatValue, c.Float)
	}
}

func Test_Parse_Float64_Valid(t *testing.T) {
	type configuration struct {
		Float float64 `key:"float"`
	}

	var floatValue float64 = 5.5
	bytes, _ := json.Marshal(floatValue)
	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(fmt.Sprintf(`{
				"float": %s
			}`, string(bytes))))).
		Parse(&c)
	if assert.NoError(t, err) {
		assert.Equal(t, floatValue, c.Float)
	}
}

func Test_Parse_Float32_Invalid(t *testing.T) {
	type configuration struct {
		Float float32 `key:"float"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"float": 5.5no float here
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Float64_Invalid(t *testing.T) {
	type configuration struct {
		Float float64 `key:"float"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"float": 5.5no float here
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Int_Invalid(t *testing.T) {
	type configuration struct {
		FieldA int `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
				"field_a": 10no int here
			}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}

func Test_Parse_Uint_Invalid(t *testing.T) {
	type configuration struct {
		FieldA uint `key:"field_a"`
	}

	var c configuration
	err := yagcl.New[configuration]().
		Add(yagcl_json.
			Source().
			Bytes([]byte(`{
			"field_a": 10no int here
		}`))).
		Parse(&c)
	assert.ErrorIs(t, err, yagcl.ErrParseValue)
}
