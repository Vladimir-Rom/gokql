package gokql

import (
	"testing"
	"time"
)

func TestMatch(t *testing.T) {
	testExpr := func(expression string, obj map[string]interface{}, expectedResult bool) {
		expr, err := Parse(expression)
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
		"a:1 or b>2",
		map[string]interface{}{
			"a": "3",
			"b": 10,
		},
		true)

	testExpr(
		"a:1 or b > '2021-05-17T01:00:00Z'",
		map[string]interface{}{
			"a": "3",
			"b": time.Now(),
		},
		true)

	testExpr(
		"a>=1",
		map[string]interface{}{
			"a": 0,
		},
		false)

	testExpr(
		"a<b",
		map[string]interface{}{
			"a": "a",
		},
		true)

	testExpr(
		"a<=b",
		map[string]interface{}{
			"a": "b",
		},
		true)

	testExpr(
		"propStr:'value1'",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExpr(
		"propStr:'value*'",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExpr(
		"propStr:value*",
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

	testExpr(
		"prop:(1 or 2)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExpr(
		"prop:(0 and 5)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		false)

	testExpr(
		"prop:(2 and 3)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExpr(
		"prop:2",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExpr(
		"prop:'a*'",
		map[string]interface{}{
			"prop": []string{"bbb", "abc", "ccc"},
		},
		true)
}

var expr *Expression
var evaluator MapEvaluator

func init() {
	var err error
	expr, err = Parse("prop1: 42")
	if err != nil {
		panic(err)
	}

	evaluator = MapEvaluator{
		map[string]interface{}{
			"prop1": 42,
		}}
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expr.Match(evaluator)
	}
}
