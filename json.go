package yagcl_json

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Bios-Marcel/yagcl"
	"github.com/buger/jsonparser"
)

// ErrNoDataSourceSpecified is thrown if none Bytes, String, Path or Reader
// of the JSONSourceSetupStepOne interface have been called.
var ErrNoDataSourceSpecified = errors.New("no data source specified; call Bytes(), String(), Reader() or Path()")

// ErrNoDataSourceSpecified is thrown if more than one of Bytes, String, Path
// or Reader of the JSONSourceSetupStepOne interface have been called.
var ErrMultipleDataSourcesSpecified = errors.New("more than one data source specified; only call one of Bytes(), String(), Reader() or Path()")

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
	// Bytes defines a string to read from directly.
	String(string) JSONSourceOptionalSetup[T]
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
	// FIXME does must actually make sense for anything but files?
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

// String implements JSONSourceSetupStepOne.String.
func (s *jsonSourceImpl) String(str string) JSONSourceOptionalSetup[*jsonSourceImpl] {
	s.bytes = []byte(str)
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
func (s *jsonSourceImpl) getBytes() (data []byte, err error) {
	// Do bytes first, since it saves us the error handling code.
	if len(s.bytes) > 0 {
		data = s.bytes
		return
	}

	// We attempt to check if the source can't be found. While we only do
	// direct file access in case a path is passed, a reader might also
	// attempt reading from a file source, therefore we try to check that
	// error as well.
	defer func() {
		if err != nil && os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist) {
			err = yagcl.ErrSourceNotFound
		}
	}()

	if s.path != "" {
		data, err = os.ReadFile(s.path)
		return
	}

	if s.reader != nil {
		if closer, ok := s.reader.(io.Closer); ok {
			defer closer.Close()
		}
		data, err = io.ReadAll(s.reader)
		return
	}

	// This should be dead code and therefore isn't covered by a test either.
	err = errors.New("verification process must have failed, please report this to the maintainer")
	return
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
func (s *jsonSourceImpl) Parse(parsingCompanion yagcl.ParsingCompanion, configurationStruct any) (bool, error) {
	if err := s.verify(); err != nil {
		return false, err
	}

	bytes, err := s.getBytes()
	if err != nil {
		if !s.must && err == yagcl.ErrSourceNotFound {
			return false, nil
		}
		return false, err
	}

	_, err = s.parse(parsingCompanion, bytes, nil, reflect.Indirect(reflect.ValueOf(configurationStruct)))
	return err == nil, err
}

func (s *jsonSourceImpl) parse(parsingCompanion yagcl.ParsingCompanion, bytes []byte, parentJsonPath []string, structValue reflect.Value) (bool, error) {
	var hasAnyFieldBeenSet bool
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		structField := structType.Field(i)
		// By default, all exported fiels are not ignored and all exported
		// fields are. Unexported fields can't be un-ignored though.
		if !parsingCompanion.IncludeField(structField) {
			continue
		}

		jsonKey, err := s.extractJSONKey(parsingCompanion, structField)
		if err != nil {
			return hasAnyFieldBeenSet, err
		}
		jsonPath := append(parentJsonPath, jsonKey)

		valueBytes, dataType, _, err := jsonparser.Get(bytes, jsonPath...)
		// Since not every field in the struct might be in the JSON, we
		// ignore these "errors".
		if err == jsonparser.KeyPathNotFoundError {
			continue
		}
		if err != nil {
			return hasAnyFieldBeenSet, newJsonparserError(jsonPath, err)
		}

		fieldType := extractNonPointerFieldType(structField.Type)
		fieldValue := structValue.Field(i)
		var value any

		// In this section we check whether custom unmarshallers are present.
		// Types with a custom unmarshaller have to be checked first before
		// attempting to parse them using default behaviour, as the behaviour
		// might differ from std/json otherwise.

		// Technically this check isn't required, as we already filter out
		// unexported fields. However, I am unsure whether this behaviour is set
		// in stone, as it hasn't been documented properly.
		// https://stackoverflow.com/questions/50279840/when-is-go-reflect-caninterface-false
		var customUnmarshalApplied bool
		if fieldValue.CanInterface() {
			newType := extractNonPointerFieldType(fieldValue.Type())
			// New pointer value, since non-pointers can't implement json.Unmarshaler.
			parsed := reflect.New(newType)
			if u, ok := parsed.Interface().(json.Unmarshaler); ok {
				// Since jsonparser strips the quotes from strings, we need to add
				// them back in order for custom unmarshalling not to fail.
				if dataType == jsonparser.String {
					// This means that strings might still contain escape sequences.
					// The implementation of UnmarshalJSON has to treat this.
					// FIXME See if this behaviour is the same in standard go json.
					valueBytes = append(append([]byte(`"`), valueBytes...), byte('"'))
				}

				if err := u.UnmarshalJSON(valueBytes); err != nil {
					return hasAnyFieldBeenSet, newUnmarshalError(jsonPath, err)
				}

				value = u
				customUnmarshalApplied = true
			} else if u, ok := parsed.Interface().(encoding.TextUnmarshaler); ok {
				// Only supported for string, as it is "TextUnmarshaler".
				if dataType == jsonparser.String {
					if err := u.UnmarshalText(valueBytes); err != nil {
						return hasAnyFieldBeenSet, newUnmarshalError(jsonPath, err)
					}

					value = u
					customUnmarshalApplied = true
				}
			}
		}

		if !customUnmarshalApplied {
			switch fieldType.Kind() {
			case reflect.String:
				if dataType != jsonparser.String {
					return hasAnyFieldBeenSet, fmt.Errorf("field '%s' had an incorrect JSON type (%s != string): %w", structField.Name, dataType.String(), yagcl.ErrParseValue)
				}
				// Can't use the raw value, as there might be escape sequences.
				// This is basically what jsonparser.GetString does.
				value, err = jsonparser.ParseString(valueBytes)
				if err != nil {
					return hasAnyFieldBeenSet, newJsonparserError(jsonPath, err)
				}
			case reflect.Struct:
				// We can't operate on any zero value, therefore we create a
				// temporary value for the struct.
				var structValue reflect.Value
				if fieldValue.IsZero() {
					structValue = reflect.New(fieldType)
				} else {
					structValue = fieldValue
				}
				structValue = reflect.Indirect(structValue)

				hasAnySubStructFieldBeenSet, err := s.parse(parsingCompanion, bytes, jsonPath, structValue)
				hasAnyFieldBeenSet = hasAnyFieldBeenSet || hasAnySubStructFieldBeenSet
				if err != nil {
					return hasAnyFieldBeenSet, err
				}

				// Only if any field of our temporary struct has been set, we
				// actually use the initialised struct for its parent.
				// Otherwise we'd initialise struct pointers that don't have a
				// single field set, losing the information of what values have
				// actually been set. Additionally, executing the rest of the
				// loop would cause a panic, as we'd try to access the value
				// that hasn't been initiliased.
				if !hasAnySubStructFieldBeenSet {
					continue
				}

				value = structValue.Interface()
			case reflect.Complex64, reflect.Complex128:
				{
					// Complex isn't supported, as for example it also isn't supported
					// by the stdlib json encoder / decoder.
					return hasAnyFieldBeenSet, fmt.Errorf("type '%s' isn't supported and won't ever be: %w", structField.Name, yagcl.ErrUnsupportedFieldType)
				}
			case reflect.Int64:
				{
					if dataType == jsonparser.String {
						if stringValue, err := jsonparser.ParseString(valueBytes); err == nil {
							// Since there are no constants for alias / struct types, we have
							// to an additional check with custom parsing, since durations
							// also contain a duration unit, such as "s" for seconds.
							if fieldType.AssignableTo(reflect.TypeOf(time.Duration(0))) {
								var errParse error
								value, errParse = time.ParseDuration(stringValue)
								if errParse != nil {
									return hasAnyFieldBeenSet, fmt.Errorf("value '%s' isn't parsable as an 'time.Duration' for field '%s': %w", stringValue, structField.Name, yagcl.ErrParseValue)
								}

								value = reflect.ValueOf(value).Convert(fieldType).Interface()
								// Parse successful, default path not needed.
								break
							}
						}
					}
				}
				// Since we seem to just have a normal int64 (or other alias type), we
				// want to proceed treating it as a normal int, which is why we
				// fallthrough.
				fallthrough
			default:
				{
					value = reflect.New(fieldType).Interface()
					err = json.Unmarshal(valueBytes, &value)
					if err != nil {
						return hasAnyFieldBeenSet, newUnmarshalError(jsonPath, err)
					}
				}
			}
		}

		hasAnyFieldBeenSet = true
		// Make sure that we have neither a pointer, not type aliased type that is incorrect.
		parsed := reflect.Indirect(reflect.ValueOf(value)).Convert(fieldType)
		if fieldValue.Kind() == reflect.Pointer {
			// Create as many values as we have pointers pointing to things.
			var pointers []reflect.Value
			lastPointer := reflect.New(fieldValue.Type().Elem())
			pointers = append(pointers, lastPointer)
			for lastPointer.Elem().Kind() == reflect.Pointer {
				lastPointer = reflect.New(lastPointer.Elem().Type().Elem())
				pointers = append(pointers, lastPointer)
			}

			pointers[len(pointers)-1].Elem().Set(parsed)
			for i := len(pointers) - 2; i >= 0; i-- {
				pointers[i].Elem().Set(pointers[i+1])
			}
			fieldValue.Set(pointers[0])
		} else {
			fieldValue.Set(parsed)
		}
	}

	return hasAnyFieldBeenSet, nil
}

func newUnmarshalError(jsonPath []string, err error) error {
	return fmt.Errorf("error unmarshalling field '%s': (%s): %w", jsonPath, err, yagcl.ErrParseValue)
}

func newJsonparserError(jsonPath []string, err error) error {
	return fmt.Errorf("error accessing json field '%s': (%s): %w", jsonPath, err, yagcl.ErrParseValue)
}

func (s *jsonSourceImpl) extractJSONKey(parsingCompanion yagcl.ParsingCompanion, structField reflect.StructField) (string, error) {
	// Custom tag
	key := structField.Tag.Get(s.KeyTag())
	if key != "" {
		return strings.Split(key, ",")[0], nil
	}

	// Fallback tag
	if key := parsingCompanion.ExtractFieldKey(structField); key != "" {
		// FIXME keyValueConverter?
		return key, nil
	}

	// No tag found
	return "", fmt.Errorf("neither tag '%s' nor the standard tag '%s' have been set for field '%s': %w", s.KeyTag(), yagcl.DefaultKeyTagName, structField.Name, yagcl.ErrExportedFieldMissingKey)
}

func extractNonPointerFieldType(fieldType reflect.Type) reflect.Type {
	if fieldType.Kind() != reflect.Pointer {
		return fieldType
	}

	return extractNonPointerFieldType(fieldType.Elem())
}
