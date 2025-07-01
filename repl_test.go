package main

import "testing"

func TestCleanInput(t *testing.T) {
	// create list of test cases

	cases := []struct {
		input    string
		expected []string
	}{
		// Basic spacing
		{
			input:    " hello  world",
			expected: []string{"hello", "world"},
		},

		// Leading/trailing whitespace
		{
			input:    "   foo bar   ",
			expected: []string{"foo", "bar"},
		},

		// Multiple spaces and tabs between words
		{
			input:    "go\tis\tawesome   ",
			expected: []string{"go", "is", "awesome"},
		},

		// Only spaces
		{
			input:    "       ",
			expected: []string{},
		},

		// Empty string
		{
			input:    "",
			expected: []string{},
		},

		// Already clean input
		{
			input:    "golang is fun",
			expected: []string{"golang", "is", "fun"},
		},

		// Mixed casing
		{
			input:    "HeLLo WoRLD",
			expected: []string{"hello", "world"},
		},

		// Newlines and tabs
		{
			input:    "hello\nworld\tgolang",
			expected: []string{"hello", "world", "golang"},
		},

		// Extra mixed whitespace
		{
			input:    " \t hello \n world \t ",
			expected: []string{"hello", "world"},
		},

		// Special characters that aren't whitespace
		{
			input:    "foo! bar?",
			expected: []string{"foo!", "bar?"},
			// Note: this test verifies that punctuation is preserved
		},

		// Words separated by multiple newlines
		{
			input:    "foo\n\n\nbar",
			expected: []string{"foo", "bar"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(actual) != len(c.expected) {
			t.Fatalf("Length mismatch.\nExpected: %v\nActual: %v", c.expected, actual)
		}

		for i := range actual {
			word := actual[i]
			excpectedWord := c.expected[i]

			if word != excpectedWord {
				t.Fatalf("Words don't match\nExpected: %v\nActual: %v", excpectedWord, word)
			}
		}
	}
}
