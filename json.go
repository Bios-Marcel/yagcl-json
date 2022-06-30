package json

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/Bios-Marcel/yagcl"
)

// ErrNoDataSourceSpecified is thrown if none Bytes, Path or Reader of the
// JSONSourceSetupStepOne interface have been called.
var ErrNoDataSourceSpecified = errors.New("no data source specified; call Bytes(), Reader() or Path()")

// ErrNoDataSourceSpecified is thrown if more than one of Bytes, Path or Reader
// of the JSONSourceSetupStepOne interface have been called.
var ErrMultipleDataSourcesSpecified = errors.New("more than one data source specified; only call one of Bytes(), Reader() or Path()")

type jsonSourceImpl struct {
	must   bool
	path   string
	bytes  []byte
	reader io.Reader
}

// JSONSourceSetupStepOne enforces the API caller to specify any data source to
// read JSON encoded data from, before being able to pass the source on to
// YAGCL.
type JSONSourceSetupStepOne[T yagcl.Source] interface {
	// Bytes defines a byte array to read from directly.
	Bytes([]byte) JSONSourceOptionalSetup[T]
	// Path defines a filepath that is accessed when YAGCL.Parse is called.
	Path(string) JSONSourceOptionalSetup[T]
	// Reader defines a reader that is accessed when YAGCL.Parse is called. IF
	// available, io.Closer.Close() is called.
	Reader(io.Reader) JSONSourceOptionalSetup[T]
}

// JSONSourceOptionalSetup offers optional Methods for configuring the source
// and exposes all methods required for a source to be passed on to YAGCL.
type JSONSourceOptionalSetup[T yagcl.Source] interface {
	yagcl.Source
	// Must declares this source as mandatory, erroring in case no data can
	// be loaded.
	// FIXME Clarify when this case happens. Only when not finding a file?
	Must() T
}

// Source creates a source for a JSON file.
func Source() JSONSourceSetupStepOne[*jsonSourceImpl] {
	return &jsonSourceImpl{}
}

// Must implements JSONSourceOptionalSetup.Must.
func (s *jsonSourceImpl) Must() *jsonSourceImpl {
	s.must = true
	return s
}

// KeyTag implements Source.Key.
func (s *jsonSourceImpl) KeyTag() string {
	return "json"
}

// Bytes implements JSONSourceSetupStepOne.Bytes.
func (s *jsonSourceImpl) Bytes(bytes []byte) JSONSourceOptionalSetup[*jsonSourceImpl] {
	s.bytes = bytes
	return s
}

// Path implements JSONSourceSetupStepOne.Path.
func (s *jsonSourceImpl) Path(path string) JSONSourceOptionalSetup[*jsonSourceImpl] {
	s.path = path
	return s
}

// Reader implements JSONSourceSetupStepOne.Reader.
func (s *jsonSourceImpl) Reader(reader io.Reader) JSONSourceOptionalSetup[*jsonSourceImpl] {
	s.reader = reader
	return s
}

// getBytes attempts to retrieve data via one of the defined data sources.
// A call to jsonSourceImpl.verify should've been done before calling this in
// order to avoid undefined behaviour.
func (s *jsonSourceImpl) getBytes() ([]byte, error) {
	if s.path != "" {
		fileData, errOpen := os.ReadFile(s.path)
		if errOpen != nil {
			if os.IsNotExist(errOpen) {
				return nil, yagcl.ErrSourceNotFound
			}
			return nil, errOpen
		}
		return fileData, nil
	}

	if len(s.bytes) > 0 {
		return s.bytes, nil
	}

	if s.reader != nil {
		closer, ok := s.reader.(io.Closer)
		if ok {
			defer closer.Close()
		}
		return io.ReadAll(s.reader)
	}

	panic("verification process must have failed, please report this to the maintainer")
}

// verify checks whether the source has been configured correctly. We attempt
// avoiding any condiguration errors by API design.
func (s *jsonSourceImpl) verify() error {
	var dataSourcesCount uint
	if s.path != "" {
		dataSourcesCount++
	}
	if len(s.bytes) > 0 {
		dataSourcesCount++
	}
	if s.reader != nil {
		dataSourcesCount++
	}

	if dataSourcesCount == 0 {
		return ErrNoDataSourceSpecified
	}
	if dataSourcesCount > 1 {
		return ErrMultipleDataSourcesSpecified
	}

	return nil
}

// Parse implements Source.Parse.
func (s *jsonSourceImpl) Parse(configurationStruct any) (bool, error) {
	if err := s.verify(); err != nil {
		return false, err
	}

	bytes, err := s.getBytes()
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(bytes, configurationStruct)
	return err == nil, err
}
