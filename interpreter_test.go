package gokql

import "testing"

func TestMatch(t *testing.T) {
	testExpr := func(expression string, obj map[string]interface{}, expectedResult bool) {
		expr, err := parse(expression)
		if err != nil {
			t.Fatal(err)
		}

		result, err := expr.Match(MapEvaluator{obj})
		if err != nil {
			t.Error(err)
		}

		if result != expectedResult {
			t.Errorf("Unexpected match result: %v", result)
		}
	}

	testExpr(
		"propStr:'value1'",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExpr(
		"propStr:'value2' or propInt:42",
		map[string]interface{}{
			"propStr": "value1",
			"propInt": 42,
		},
		true)

	testExpr(
		"propStr:'value2' or not propInt:42",
		map[string]interface{}{
			"propStr": "value1",
			"propInt": 42,
		},
		false)

	testExpr(
		"propStr:'value2' or nested:{int:13}",
		map[string]interface{}{
			"propStr": "value1",
			"nested": map[string]interface{}{
				"int": 13,
			},
		},
		true)

	testExpr(
		"propStr:('value1' or value2)",
		map[string]interface{}{
			"propStr": "value2",
		},
		true)
}
