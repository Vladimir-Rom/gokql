package gokql

import (
	"testing"
	"time"
)

func TestBasicMatch(t *testing.T) {
	testExprMap(
		t,
		"a:1 or b>2",
		map[string]interface{}{
			"a": "3",
			"b": 10,
		},
		true)

	testExprMap(
		t,
		"a:1 or b > '2021-05-17T01:00:00Z'",
		map[string]interface{}{
			"a": "3",
			"b": time.Now(),
		},
		true)

	testExprMap(
		t,
		"a>=1",
		map[string]interface{}{
			"a": 0,
		},
		false)

	testExprMap(
		t,
		"a<b",
		map[string]interface{}{
			"a": "a",
		},
		true)

	testExprMap(
		t,
		"a<=b",
		map[string]interface{}{
			"a": "b",
		},
		true)

	testExprMap(
		t,
		"propStr:'value1'",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExprMap(
		t,
		"propStr:'value*'",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExprMap(
		t,
		"propStr:value*",
		map[string]interface{}{
			"propStr": "value1",
		},
		true)

	testExprMap(
		t,
		"propStr:'value2' or propInt:42",
		map[string]interface{}{
			"propStr": "value1",
			"propInt": 42,
		},
		true)

	testExprMap(
		t,
		"propStr:'value2' or not propInt:42",
		map[string]interface{}{
			"propStr": "value1",
			"propInt": 42,
		},
		false)

	testExprMap(
		t,
		"propStr:'value2' or nested:{int:13}",
		map[string]interface{}{
			"propStr": "value1",
			"nested": map[string]interface{}{
				"int": 13,
			},
		},
		true)

	testExprMap(
		t,
		"propStr:('value1' or value2)",
		map[string]interface{}{
			"propStr": "value2",
		},
		true)

	testExprMap(
		t,
		"prop:(1 or 2)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExprMap(
		t,
		"prop:(0 and 5)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		false)

	testExprMap(
		t,
		"prop:(2 and 3)",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExprMap(
		t,
		"prop:2",
		map[string]interface{}{
			"prop": []int{0, 2, 3},
		},
		true)

	testExprMap(
		t,
		"prop:'a*'",
		map[string]interface{}{
			"prop": []string{"bbb", "abc", "ccc"},
		},
		true)
}

func TestReflectMatch(t *testing.T) {
	type nested struct{ NestedProp string }
	testExpr(
		t,
		"Prop:val and Nested:{NestedProp:val2}",
		NewReflectEvaluator(
			struct {
				Prop   string
				Nested nested
			}{"val", nested{"val2"}}),
		true)
}

func testExprMap(t *testing.T, expression string, obj map[string]interface{}, expectedResult bool) {
	testExpr(t, expression, MapEvaluator{obj}, expectedResult)
}

func testExpr(t *testing.T, expression string, evaluator Evaluator, expectedResult bool) {
	expr, err := Parse(expression)
	if err != nil {
		t.Fatal(err)
	}

	result, err := expr.Match(evaluator)
	if err != nil {
		t.Error(err)
	}

	if result != expectedResult {
		t.Errorf("Unexpected match result: %v for expression %s", result, expression)
	}
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
	expr.Match(evaluator)
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expr.Match(evaluator)
	}
}
