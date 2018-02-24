package envconf

import (
	"fmt"
	"os"
	"testing"
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
				t.Errorf("For %s, Wrong error: %s, expect %s", str, expectErr)
			}

		}
	}

}

func TestParse(t *testing.T) {
	s := struct {
		Simple  string   `env:"T_SIMPLE"`
		Default string   `env:"T_DEFAULT" default:"default"`
		Bytes   Base64   `env:"T_SECRET"`
		Slice   []string `env:"T_SLICE"`
	}{}

	os.Setenv("T_SIMPLE", "simple")
	os.Setenv("T_SECRET", "dGVzdCBzdHJpbmc=")
	os.Setenv("T_SLICE", "val1, val2")
	if err := Parse(&s); err != nil {
		t.Fatal(err.Error())
	}

	if s.Simple != "simple" {
		t.Errorf("Expect 'simple' got '%s'", s.Simple)
	}

	if s.Default != "default" {
		t.Errorf("Expect 'default' got '%s'", s.Simple)
	}

	if string(s.Bytes) != "test string" {
		t.Errorf("Expected 'test string' in bytes, got %s", string(s.Bytes))
	}
	if len(s.Slice) != 2 {
		t.Errorf("Expected 2 elements, got %v", s.Slice)
	} else if s.Slice[0] != "val1" || s.Slice[1] != "val2" {
		t.Errorf("Expected the values, got %v", s.Slice)
	}

}
