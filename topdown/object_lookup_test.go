package topdown

import (
	"fmt"
	"testing"
)

func TestObjectLookup(t *testing.T) {
	cases := []struct {
		note     string
		object   string
		key      interface{}
		fallback interface{}
		expected interface{}
	}{
		{
			note:     "basic case . found",
			object:   `{"a": "b"}`,
			key:      `"a"`,
			fallback: `"c"`,
			expected: `"b"`,
		},
		{
			note:     "basic case . not found",
			object:   `{"a": "b"}`,
			key:      `"c"`,
			fallback: `"c"`,
			expected: `"c"`,
		},
		{
			note:     "complex value . found",
			object:   `{"a": {"b": "c"}}`,
			key:      `"a"`,
			fallback: "true",
			expected: `{"b": "c"}`,
		},
		{
			note:     "complex value . not found",
			object:   `{"a": {"b": "c"}}`,
			key:      `"b"`,
			fallback: "true",
			expected: "true",
		},
		{
			note:     "lookup value . found: exact path",
			object:   `{"a": {"b": "x"}}`,
			key:      `"a.b"`,
			fallback: "true",
			expected: `"x"`,
		},
		{
			note:     "lookup value . not found: over-reaching path",
			object:   `{"a": {"b": "x"}}`,
			key:      `"a.b.c.d"`,
			fallback: "true",
			expected: `true`,
		},
		{
			note:     "lookup value . found: w/index",
			object:   `{"a": [{"b": ["x"]}]}`,
			key:      `"a.0.b.0"`,
			fallback: "true",
			expected: `"x"`,
		},
		{
			note:     "lookup value . not found: bad path",
			object:   `{"a": {"b": "x"}}`,
			key:      `"b.c"`,
			fallback: "true",
			expected: "true",
		},
		{
			note:     "lookup value . not found: bad over-reaching path",
			object:   `{"a": {"b": "x"}}`,
			key:      `"b.c.x.y.z"`,
			fallback: "true",
			expected: "true",
		},
		{
			note:     "lookup value . not found: lookup into empty object",
			object:   `{}`,
			key:      `"b.c"`,
			fallback: "true",
			expected: "true",
		},
	}

	for _, tc := range cases {
		rules := []string{
			fmt.Sprintf("p = x { x := object.lookup(%s, %s, %s) }", tc.object, tc.key, tc.fallback),
		}
		runTopDownTestCase(t, map[string]interface{}{}, tc.note, rules, tc.expected)
	}
}
