package envconf

import (
	"fmt"
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
