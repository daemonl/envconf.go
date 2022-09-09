// Package envconf parses environment variables into structs.
// It supports multiple types, however the core type is always a string.
//
// Translators are available which manipulate the string value before setting,
// e.g. `!base64:SGVsbG8gV29ybGQ=` will cast to a string as "Hello World".
//
// There is no specific handling for bytes when using this method, it is handled
// as a string entirely, if you are expecting actual bytes use a type like Hex
// or Base64 which will directly translate the env var string to bytes
//
// Standard conversion from string to int, bool etc work, as well as custom
// types which satisfy `SetterFromEnv` (on a pointer, like JSON)
//
// Combining translators and custom types is perfectly fine. The string
// translations will happen first, then the output will be passed into
// FromEnvString
package envconf

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Translator is responsible for taking a string and converting it to the
// output string, either by looking it up in an external service, decrypring,
// or anything else you can think of
type Translator interface {
	Translate(in string) (string, error)
}

// SetterFromEnv is used by SetFromString for custom types
type SetterFromEnv interface {
	FromEnvString(string) error
}

// Base64 is a byte array which is loaded from a URL base64 string
type Base64 []byte

// FromEnvString Satisfies SetterFromEnv
func (b64 *Base64) FromEnvString(in string) error {
	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return err
	}
	*b64 = b
	return nil
}

// Hex is a byte array which is loaded from a Hex string
type Hex []byte

// FromEnvString Satisfies SetterFromEnv
func (h *Hex) FromEnvString(in string) error {
	b, err := hex.DecodeString(in)
	if err != nil {
		return err
	}
	*h = b
	return nil
}

// TranslatorFunc is an adaptor to allow the use of ordinary functions as Translators
type TranslatorFunc func(in string) (string, error)

// Translate satisfies the Translator interface
func (tf TranslatorFunc) Translate(in string) (string, error) {
	return tf(in)
}

// Parser holds a list of Translator functions
type Parser struct {
	Translators map[string]Translator
}

// New returns a new Parser with an empty translator set
func New() *Parser {
	return &Parser{
		Translators: map[string]Translator{},
	}
}

// RegisterTranslatorFunc adds a translator function to the list of
// translators. It replaces any existing function with the given name
func (p *Parser) RegisterTranslatorFunc(name string, translator func(string) (string, error)) {
	p.Translators[name] = TranslatorFunc(translator)
}

// Parse reads the tags of dest to set any fields which should be parsed from
// the environment. The `env` tag gives the name of the variable. If the
// environment variable evaluates to an empty string, the value of `default` is
// used, or an error is thrown if the `default` tag is omitted.
// To allow optional parameters, set default to an empty string
func (p Parser) Parse(dest interface{}) error {

	rt := reflect.TypeOf(dest).Elem()
	rv := reflect.ValueOf(dest).Elem()
	for i := 0; i < rv.NumField(); i++ {
		tag := rt.Field(i).Tag
		envName := tag.Get("env")
		if envName == "" {
			continue
		}
		envVal := os.Getenv(envName)
		if envVal == "" {
			if defaultValue, ok := tag.Lookup("default"); ok {
				envVal = defaultValue
			} else if req, ok := tag.Lookup("required"); ok && req == "false" {
				continue
			} else {
				return fmt.Errorf("Required ENV var not set: %v", tag)
			}
		}

		envVal, err := p.Translate(envVal)
		if err != nil {
			return fmt.Errorf("In field %s: %s", envName, err)
		}

		fieldVal := rv.Field(i)

		fieldInterface := fieldVal.Addr().Interface()

		actualType := fieldVal.Kind()
		if actualType == reflect.Pointer {
			elemType := fieldVal.Type().Elem()
			newVal := reflect.New(elemType)
			fieldVal.Set(newVal)
			fieldVal = newVal
			actualType = fieldVal.Elem().Kind()
		}

		if actualType == reflect.Struct {
			if !strings.HasPrefix(envVal, "{") {
				return fmt.Errorf("In field %s: struct fields should be set using JSON strings", envName)
			}

			if err := json.Unmarshal([]byte(envVal), fieldInterface); err != nil {
				return err
			}
			continue
		}

		if err := SetFromString(fieldInterface, envVal); err != nil {
			return fmt.Errorf("In field %s: %s", envName, err)
		}

	}
	return nil
}

// SetFromString attempts to translate a string to the given interface. Must be a pointer.
// Standard Types string, bool, int, int(8-64) float(32, 64) and []string.
// Custom types must have method FromEnvString(string) error
func SetFromString(fieldInterface interface{}, stringVal string) error {

	if withSetter, ok := fieldInterface.(SetterFromEnv); ok {
		return withSetter.FromEnvString(stringVal)
	}

	var err error

	switch field := fieldInterface.(type) {
	case *string:
		*field = stringVal
		return nil
	case *bool:
		bVal := strings.HasPrefix(strings.ToLower(stringVal), "t")
		*field = bVal
		return nil

	case *int:
		*field, err = strconv.Atoi(stringVal)
		return err
	case *int64:
		*field, err = strconv.ParseInt(stringVal, 10, 64)
		return err
	case *int32:
		field64, err := strconv.ParseInt(stringVal, 10, 32)
		*field = int32(field64)
		return err
	case *int16:
		field64, err := strconv.ParseInt(stringVal, 10, 16)
		*field = int16(field64)
		return err
	case *int8:
		field64, err := strconv.ParseInt(stringVal, 10, 8)
		*field = int8(field64)
		return err

	case *uint:
		field64, err := strconv.ParseUint(stringVal, 10, 64)
		*field = uint(field64)
		return err
	case *uint64:
		*field, err = strconv.ParseUint(stringVal, 10, 64)
		return err
	case *uint32:
		field64, err := strconv.ParseUint(stringVal, 10, 32)
		*field = uint32(field64)
		return err
	case *uint16:
		field64, err := strconv.ParseUint(stringVal, 10, 16)
		*field = uint16(field64)
		return err
	case *uint8:
		field64, err := strconv.ParseUint(stringVal, 10, 8)
		*field = uint8(field64)
		return err

	case *float64:
		*field, err = strconv.ParseFloat(stringVal, 64)
		return err
	case *float32:
		field64, err := strconv.ParseFloat(stringVal, 32)
		*field = float32(field64)
		return err

	// TODO: Support an array of anything. Using reflect?
	case *[]string:
		vals := strings.Split(stringVal, ",")
		out := make([]string, 0, len(vals))
		for _, val := range vals {
			stripped := strings.TrimSpace(val)
			if stripped == "" {
				continue
			}
			out = append(out, stripped)
		}
		*field = out
		return nil
	}

	return fmt.Errorf("unsupported type %T", fieldInterface)
}

var reTranslate = regexp.MustCompile(`^!([a-zA-Z0-9_\-]+):`)

// Translate runs the parser's translators on the string
func (p Parser) Translate(val string) (string, error) {
	if strings.HasPrefix(val, `\!`) {
		val = "!" + val[2:]
		return val, nil
	}

	match := reTranslate.FindStringSubmatch(val)
	if len(match) != 2 {
		return val, nil
	}
	name := match[1]
	input := val[len(match[1])+2:]
	if trans, ok := p.Translators[name]; ok {
		out, err := trans.Translate(input)
		if err != nil {
			return out, err
		}
		return p.Translate(out)
	}
	return "", fmt.Errorf("no translator named %s", name)

}

// DefaultParser is used by the Parse and Translate funcs for convenience
var DefaultParser = Parser{
	Translators: map[string]Translator{
		"base64": TranslatorFunc(Base64Translator),
	},
}

// Parse uses DefaultParser.Parse
func Parse(dest interface{}) error {
	return DefaultParser.Parse(dest)
}

// Translate uses the DefaultParser.Translate
func Translate(in string) (string, error) {
	return DefaultParser.Translate(in)
}

// Base64Translator decodes a URL Base64 encoded string. Not safe for byte
// data, use the Base64 type instead
func Base64Translator(in string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
