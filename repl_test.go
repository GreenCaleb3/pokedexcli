package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := [] struct {
		input string
		expected []string
	}{
		{
			input: " hello world ",
			expected: []string{"hello", "world"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		
		if (len(actual) != len(c.expected)) {
			t.Errorf("Actual slice different length than expected")
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if (word != expectedWord) {
				t.Errorf("Actual word does not match expected")
			}
		}
	}
}
