package envconf

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// Translator is responsible for taking a string and converting it to the
// output string, either by looking it up in an external service, decrypring,
// or anything else you can think of
type Translator interface {
	Translate(in string) (string, error)
}

type setterFromEnv interface {
	FromEnvString(string) error
}

type Parser struct {
	Translators map[string]Translator
}

func New() *Parser {
	return &Parser{
		Translators: map[string]Translator{},
	}
}

type Base64 []byte

func (b64 *Base64) FromEnvString(in string) error {
	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return err
	}
	*b64 = b
	return nil
}

type Hex []byte

func (h *Hex) FromEnvString(in string) error {
	b, err := hex.DecodeString(in)
	if err != nil {
		return err
	}
	*h = b
	return nil
}

type TranslatorFunc func(in string) (string, error)

func (tf TranslatorFunc) Translate(in string) (string, error) {
	return tf(in)
}

func (p *Parser) RegisterTranslatorFunc(name string, translator func(string) (string, error)) {
	p.Translators[name] = TranslatorFunc(translator)
}

// Parse reads the tags of dest to set any fields which should be parsed from
// the environment. the `env` tag gives the name of the variable. If the
// environment variable evaluates to an empty string, the value of `default` is
// used, or an error is thrown if the `default` tag is omitted.
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
			} else {
				return fmt.Errorf("Required ENV var not set: %v", tag)
			}
		}

		envVal, err := p.Translate(envVal)
		if err != nil {
			return err
		}

		fieldInterface := rv.Field(i).Addr().Interface()
		if withSetter, ok := fieldInterface.(setterFromEnv); ok {
			withSetter.FromEnvString(envVal)
			continue
		}

		switch field := fieldInterface.(type) {
		case *string:
			*field = envVal
		case *bool:
			bVal := strings.HasPrefix(strings.ToLower(envVal), "t")
			*field = bVal

		case *[]string:
			vals := strings.Split(envVal, ",")
			out := make([]string, 0, len(vals))
			for _, val := range vals {
				out = append(out, strings.TrimSpace(val))
			}
			*field = out
		default:
			return fmt.Errorf("Values of type %T are not supported", field)
		}
	}
	return nil
}

var reTranslate = regexp.MustCompile(`^!([a-zA-Z0-9_\-]+):`)

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
	} else {
		return "", fmt.Errorf("no translator named %s", name)
	}

}

var DefaultParser = Parser{
	Translators: map[string]Translator{
		"base64": TranslatorFunc(Base64Translator),
	},
}

func Parse(dest interface{}) error {
	return DefaultParser.Parse(dest)
}

func Translate(in string) (string, error) {
	return DefaultParser.Translate(in)
}

func Base64Translator(in string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
