package gokql

import "testing"

func TestParse(t *testing.T) {
	testExpr := func(expression string, expectedExpr string) {
		expr, err := parse(expression)
		if err != nil {
			t.Fatal(err)
		}

		if es := expr.String(); es != expectedExpr {
			t.Errorf("Wrong reconstructed expression: " + es + ". Expected: " + expectedExpr)
		}
	}

	testExpr("a:'1'", "a:1")
	testExpr("a:c or b:2", "(a:c or b:2)")
	testExpr("a:c or b:2 and c:3", "(a:c or (b:2 and c:3))")
	testExpr("(a:c or b:2) and c:3", "((a:c or b:2) and c:3)")
	testExpr(
		"a.b:c or b:2 and (c:3 or d:{da:a or db:'b'}) or list:(1 or 2 or 3)",
		"(a.b:c or (b:2 and (c:3 or d:{(da:a or db:b)})) or list:(1 or 2 or 3))")
}
