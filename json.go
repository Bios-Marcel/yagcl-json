package json

import (
	"encoding/json"
	"os"

	"github.com/Bios-Marcel/yagcl"
)

// DO NOT CREATE INSTANCES MANUALLY, THIS IS ONLY PUBLIC IN ORDER FOR GODOC
// TO RENDER AVAILABLE FUNCTIONS.
type JSONSource struct {
	must bool
	path string
}

// Source creates a source for a JSON file.
func Source(path string) *JSONSource {
	return &JSONSource{path: path}
}

// Must indicates that this Source will return an error during parsing if no
// parsable data can be found.
func (s *JSONSource) Must() *JSONSource {
	s.must = true
	return s
}

// KeyTag implements Source.Key.
func (s *JSONSource) KeyTag() string {
	return "json"
}

// Parse implements Source.Parse.
func (s *JSONSource) Parse(configurationStruct any) (bool, error) {
	file, errOpen := os.OpenFile(s.path, os.O_RDONLY, os.ModePerm)
	if os.IsNotExist(errOpen) {
		if s.must {
			return false, yagcl.ErrSourceNotFound
		}
		return false, nil
	}

	err := json.
		NewDecoder(file).
		Decode(configurationStruct)
	return err != nil, err
}
