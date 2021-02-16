package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestQueryInterpolation(t *testing.T) {
	cases := []struct {
		q                query
		vars             map[string]string
		expectedQueryStr string
	}{
		{
			q:                query{"", ""},
			vars:             map[string]string{},
			expectedQueryStr: "",
		},
		{
			q:                query{"", "no vars"},
			vars:             nil,
			expectedQueryStr: "no vars",
		},
		{
			q:                query{"", "{{var1}}"},
			vars:             map[string]string{},
			expectedQueryStr: "{{var1}}",
		},
		{
			q:                query{"", "{{var1}} and {{var2}}."},
			vars:             map[string]string{"var1": "foo", "var2": "bar"},
			expectedQueryStr: "foo and bar.",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			c.q.interpolate(c.vars)
			if !reflect.DeepEqual(c.expectedQueryStr, c.q.Query) {
				fmt.Printf("wanted:\n%+v\ngot:\n%+v\n", c.expectedQueryStr, c.q.Query)
				t.Fail()
			}
		})
	}
}
