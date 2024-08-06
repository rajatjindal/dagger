package strcase

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToPascal(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{input: "a-string", expected: "AString"},
		{input: "AString", expected: "AString"},
		{input: "hello world", expected: "HelloWorld"},
		{input: "this is a test", expected: "ThisIsATest"},
		{input: "ToPascal Function", expected: "ToPascalFunction"},
		{input: "word", expected: "Word"},
		{input: "Word", expected: "Word"},
		{input: "WORD", expected: "Word"},
		{input: " hello world ", expected: "HelloWorld"},
		{input: "   multiple   spaces  ", expected: "MultipleSpaces"},
		{input: "hello-world", expected: "HelloWorld"},
		{input: "hello_world", expected: "HelloWorld"},
		{input: "hello.world", expected: "HelloWorld"},
		{input: "hello&world", expected: "Hello&world"},
		{input: "hello2world", expected: "Hello2world"},
		{input: "2024 year", expected: "2024Year"},
		{input: "this is 4 you", expected: "ThisIs4You"},
		{input: "héllo wörld", expected: "HélloWörld"},
		{input: "a--b--c", expected: "ABC"},
		{input: "a_b_c", expected: "ABC"},
		{input: "a.b.c", expected: "ABC"},
		{input: "a&b&c", expected: "A&b&c"},
		{input: "a", expected: "A"},
		{input: "A", expected: "A"},
		{input: "a b", expected: "AB"},
		{input: " a ", expected: "A"},
		{input: " ", expected: ""},
		{input: "    ", expected: ""},
		{input: "", expected: ""},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			output := ToPascal(tc.input)
			require.Equal(t, tc.expected, output, "input: %q", tc.input)
		})
	}
}

func TestToCamel(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{input: "a-string", expected: "aString"},
		{input: "AString", expected: "aString"},
		{input: "hello world", expected: "helloWorld"},
		{input: "this is a test", expected: "thisIsATest"},
		{input: "ToCamelCase function", expected: "toCamelCaseFunction"},
		{input: "word", expected: "word"},
		{input: "Word", expected: "word"},
		{input: "WORD", expected: "word"},
		{input: " hello world ", expected: "helloWorld"},
		{input: "   multiple   spaces  ", expected: "multipleSpaces"},
		{input: "hello-world", expected: "helloWorld"},
		{input: "hello_world", expected: "helloWorld"},
		{input: "hello.world", expected: "helloWorld"},
		{input: "hello&world", expected: "hello&world"},
		{input: "hello2world", expected: "hello2world"},
		{input: "2024 year", expected: "2024Year"},
		{input: "this is 4 you", expected: "thisIs4You"},
		{input: "héllo wörld", expected: "hélloWörld"},
		{input: "a--b--c", expected: "aBC"},
		{input: "a_b_c", expected: "aBC"},
		{input: "a.b.c", expected: "aBC"},
		{input: "a&b&c", expected: "a&b&c"},
		{input: "a", expected: "a"},
		{input: "A", expected: "a"},
		{input: "a b", expected: "aB"},
		{input: " a ", expected: "a"},
		{input: " ", expected: ""},
		{input: "    ", expected: ""},
		{input: "", expected: ""},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			output := ToCamel(tc.input)
			require.Equal(t, tc.expected, output, "input: %q", tc.input)
		})
	}
}
