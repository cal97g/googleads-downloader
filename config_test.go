package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTemplateSubstitution(t *testing.T) {
	cases := []struct {
		Vars        map[string]string
		Expected    map[string]string
		ExpectedErr error
	}{
		{
			Vars: map[string]string{
				"foo": "bar",
				"bar":  "foo",
			},
			Expected: map[string]string{
				"foo": "bar",
				"bar":  "foo",
			},
		},
		{
			Vars: map[string]string{
				"foo": "FOO",
				"bar": "BAR",
				"kebab":  "{{Concat(foo,bar)}}",
			},
			Expected: map[string]string{
				"foo": "FOO",
				"bar": "BAR",
				"kebab":  "FOOBAR",
			},
		},
		{
			Vars: map[string]string{
				"i": "I",
				"love": "LOVE",
				"kebab":  "{{Concat(i,love,KEBAB)}}",
				"alot":  "{{Concat(  kebab   ,   ALOT  )}}",
			},
			Expected: map[string]string{
				"i": "I",
				"love": "LOVE",
				"kebab":  "ILOVEKEBAB",
				"alot":  "ILOVEKEBABALOT",
			},
		},
		{
			Vars: map[string]string{
				"end_date": "2019-11-12",
				"period": "10",
				"end_date-period":  "{{DateSubDays(end_date,period)}}",
			},
			Expected: map[string]string{
				"end_date": "2019-11-12",
				"period": "10",
				"end_date-period":  "2019-11-02",
			},
		},
	}

	for _, test := range cases {
		c := config{TemplateVars: test.Vars}

		err := c.ParseTemplateVars()
		if err != test.ExpectedErr {
			fmt.Println(test.Vars)
			fmt.Printf("unexpected error: %v", err)
			t.Fail()
		}

		if !reflect.DeepEqual(c.TemplateVars, test.Expected) {
			fmt.Printf("wanted:\n%+v\ngot:\n%+v\n", test.Expected, c.TemplateVars)
			t.Fail()
		}
	}
}
