package gokql

import "testing"

func TestWildcardMatch(t *testing.T) {
	test := func(str string, wildcard string, expected bool) {
		w := newWildcard(wildcard)
		res := w.Match(str)
		if res != expected {
			t.Fatalf("Unexpected match result for string '%s' by wildcard '%s': %v. Expected: %v", str, wildcard, res, expected)
		}
	}

	test("asd", "", true)
	test("asd", "*eee*", false)
	test("asd", "*eee*", false)
	test("asd", "*asd*", true)
	test("", "*", true)
	test("asd", "a**d", true)
	test("asd", "a*d", true)
	test("asd", "asc", false)
	test("asd", "asd", true)
	test("asd", "asd*", true)
	test("asd", "*asd", true)
	test("asd", "*", true)
	test("aaa-bbbccc", "aaa*bbb", false)
	test("aaa-bbbccc", "aaa*bbb*", true)
	test("aaa-bbbccc", "aaa*bbb*c", true)
}
