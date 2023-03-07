package envconf

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTranslate(t *testing.T) {

	p := New()
	p.RegisterTranslatorFunc("hello", func(in string) (string, error) {
		return in + "-world", nil
	})
	p.RegisterTranslatorFunc("err", func(in string) (string, error) {
		return "", fmt.Errorf("ERR: %s", in)
	})

	for input, expect := range map[string]string{
		"foo":               "foo",
		`\!foo`:             "!foo",
		`\!foo:`:            "!foo:",
		"!hello:foo":        "foo-world",
		"!hello:!hello:bar": "bar-world-world",
		"not!foo":           "not!foo",
		"!notbar":           "!notbar",
	} {

		output, err := p.Translate(input)
		if err != nil {
			t.Fatal(err.Error())
		}
		if output != expect {
			t.Errorf("For %s Got %s, Expect %s", input, output, expect)
		}
	}

	for input, expectErr := range map[string]string{
		"!asdf:thing": "no translator named asdf",
		"!err:thing":  "ERR: thing",
	} {
		_, err := p.Translate(input)
		if err == nil {
			t.Errorf("For %s, Got no error", input)
		} else {
			str := err.Error()
			if str != expectErr {
				t.Errorf("For %s, Wrong error: %s, expect %s", input, str, expectErr)
			}

		}
	}

}

type TestJSONStruct struct {
	Foo string `json:"foo"`
	Bar uint64 `json:"bar"`
}

func TestParse(t *testing.T) {

	s := struct {
		Simple                   string   `env:"T_SIMPLE"`
		Default                  string   `env:"T_DEFAULT" default:"default"`
		Bytes64                  Base64   `env:"T_BYTES64"`
		BytesHex                 Hex      `env:"T_BYTESHEX" default:"48656C6C6F20576F726C64"`
		Slice                    []string `env:"T_SLICE"`
		Bool                     bool     `env:"T_BOOL"`
		Translate                string   `env:"T_TRANSLATE"`
		IgnoreExported           string
		ignorePrivate            string
		JSONStruct               TestJSONStruct  `env:"T_JSON_0"`
		OptionalJSONStructSet    *TestJSONStruct `env:"T_JSON_1" required:"false"`
		OptionalJSONStructNotSet *TestJSONStruct `env:"T_JSON_2" required:"false"`
	}{}

	os.Setenv("T_SIMPLE", "simple")
	os.Setenv("T_BYTES64", "dGVzdCBzdHJpbmc=")
	os.Setenv("T_TRANSLATE", "!base64:dGVzdCBzdHJpbmc=")
	os.Setenv("T_SLICE", "val1, val2")
	os.Setenv("T_BOOL", "true")
	os.Setenv("T_JSON_0", `{"foo":"fooval","bar":100}`)
	os.Setenv("T_JSON_1", `{"foo":"fooval","bar":100}`)
	if err := Parse(&s); err != nil {
		t.Fatal(err.Error())
	}

	if s.Simple != "simple" {
		t.Errorf("Expect 'simple' got '%s'", s.Simple)
	}

	if s.Default != "default" {
		t.Errorf("Expect 'default' got '%s'", s.Simple)
	}

	if strVal := string(s.Bytes64); strVal != "test string" {
		t.Errorf("Expected 'test string' in bytes, got %s", strVal)
	}

	if strVal := string(s.Translate); strVal != "test string" {
		t.Errorf("Expected 'test string' in bytes, got %s", strVal)
	}

	if len(s.Slice) != 2 {
		t.Errorf("Expected 2 elements, got %v", s.Slice)
	} else if s.Slice[0] != "val1" || s.Slice[1] != "val2" {
		t.Errorf("Expected the values, got %v", s.Slice)
	}

	if strVal := string(s.BytesHex); strVal != "Hello World" {
		t.Errorf("Bad Hex Decoding: %s", strVal)
	}

	if !s.Bool {
		t.Errorf("Should be true")
	}

	if s.JSONStruct.Foo != "fooval" || s.JSONStruct.Bar != 100 {
		t.Errorf("T_JSON_0 broke: %v", s.JSONStruct)
	}

	if s.OptionalJSONStructNotSet != nil {
		t.Errorf("T_JSON_2 should be nil")
	}
	if s.OptionalJSONStructSet == nil {
		t.Errorf("T_JSON_1 should not be nil")
	} else if s.OptionalJSONStructSet.Foo != "fooval" || s.OptionalJSONStructSet.Bar != 100 {
		t.Errorf("T_JSON_0 broke: %v", s.OptionalJSONStructSet)
	}
}

func TestSadNotSet(t *testing.T) {
	s := struct {
		Simple string `env:"T_NOT_SET"`
	}{}

	if err := Parse(&s); err == nil {
		t.Errorf("Expected Error")
	} else {
		msg := err.Error()
		if !strings.Contains(msg, "not set") ||
			!strings.Contains(msg, "T_NOT_SET") {
			t.Errorf("Didn't match rules: %s", msg)
		}
	}
}

func TestSadUnsupportedType(t *testing.T) {
	s := struct {
		Simple []byte `env:"T_NOT_SET" default:"value"`
	}{}

	if err := Parse(&s); err == nil {
		t.Errorf("Expected Error")
	} else {
		msg := err.Error()
		if !strings.Contains(msg, "unsupported type") ||
			!strings.Contains(msg, "T_NOT_SET") {
			t.Errorf("Didn't match rules: %s", msg)
		}
	}
}

func TestSadBadEncoding(t *testing.T) {
	s := struct {
		Simple testing.T `env:"T_NOT_SET" default:"!base64:In==Valid"`
	}{}

	if err := Parse(&s); err == nil {
		t.Errorf("Expected Error")
	} else {
		msg := err.Error()
		if !strings.Contains(msg, "base64") ||
			!strings.Contains(msg, "T_NOT_SET") {
			t.Errorf("Didn't match rules: %s", msg)
		}
	}
}

func TestTypes(t *testing.T) {

	for in, expect := range map[string]interface{}{
		"1": int64(1),
		"2": int32(2),
		"3": int16(3),
		"4": int8(4),
		"5": int(5),

		"String": string("String"),

		"True":  true,
		"False": false,

		"1.1": float64(1.1),
		"1.2": float32(1.2),

		"10": float64(10),
		"11": float32(11),

		"1h": time.Duration(time.Hour),
	} {

		gotVal := reflect.New(reflect.TypeOf(expect))
		got := gotVal.Interface()

		err := SetFromString(got, in)
		if err != nil {
			t.Errorf("%s: %s", in, err.Error())
			continue
		}

		gotRaw := gotVal.Elem().Interface()
		if !reflect.DeepEqual(expect, gotRaw) {
			t.Errorf("Was Not Equal %s. \n%#v \n%#v", in, expect, gotRaw)
			continue
		}

	}
}
