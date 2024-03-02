package gokql

import (
	"testing"
	"time"
)

type testStruct struct {
	Pint      int
	Pint8     int8
	Pint16    int16
	Pint32    int32
	Puint     uint
	Puint8    uint8
	Puint16   uint16
	Puint32   uint32
	Pfloat32  float32
	Pstring   string
	Pint64    int64
	Puint64   uint64
	Pfloat64  float64
	Pbool     bool
	Ptime     time.Time
	Pduration time.Duration
}

func TestBasicMatch(t *testing.T) {
	testExprMap(
		t,
		"propStr:('value1' or value2)",
		map[string]interface{}{
			"propStr": "value3",
		},
		false)

	testExprMap(
		t,
		"a:1 or b>2",
		map[string]interface{}{
			"a": "3",
			"c": 10,
		},
		false)

	testExprMap(
		t,
		"a:3 or b>2",
		map[string]interface{}{
			"a": "3",
			"c": 10,
		},
		true)

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
		"propStr:('value1' or value2)",
		map[string]interface{}{
			"propStr": "value2",
		},
		true)
}

func TestArrays(t *testing.T) {
	obj := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"prop": "value",
				"arr": []map[string]interface{}{
					{
						"prop1": "val1",
						"prop2": "val2",
						"prop3": "val3",
					},
				},
				"arr2": []any{
					map[string]any{
						"prop1": "val1",
					},
				},
			},
		},
	}
	testExprMap(t, "level1.level2.arr2:{prop1:val1}", obj, true)
	testExprMap(t, "level1.level2:{prop:value and arr:{prop1:val1 and prop2:val2}}", obj, true)
	testExprMap(t, "level1.notexisted.arr:{prop1:val1 and prop2:val2}", obj, false)
	testExprMap(t, "level1.level2.arr:{prop1:val1 and prop2:val2}", obj, true)
	testExprMap(t, "level1.level2.arr:{prop1:val1 and prop2:val3}", obj, false)
	testExprMap(t, "level1.level2.arr:{prop1:val1 or prop2:val3}", obj, true)

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
	type nested struct {
		NestedProp  string
		NestedProp2 string
	}
	obj := struct {
		Prop        string
		Nested      nested
		NestedSlice []nested
	}{
		"val",
		nested{"val2", ""},
		[]nested{{"val1_1", "val2_1"}, {"val1_2", "val2_2"}}}

	testExpr(t, "NestedSlice:{NestedProp:val1_1}", NewReflectEvaluator(obj), true)
	testExpr(t, "NestedSlice:{NestedProp:val1_2}", NewReflectEvaluator(obj), true)
	testExpr(t, "NestedSlice:{NestedProp:val1_3}", NewReflectEvaluator(obj), false)
	testExpr(t, "NestedSlice:{NestedProp:(val1_3 or val1_1)}", NewReflectEvaluator(obj), true)
	testExpr(t, "NestedSlice:{NestedProp:val1_2 and NestedProp2:val2_2}", NewReflectEvaluator(obj), true)

	testExpr(
		t,
		"Prop:val and Nested:{NestedProp:val2} or NotExisted:{Foo:Bar}",
		NewReflectEvaluator(obj),
		true)

	testExpr(
		t,
		"Prop:val and notexisted:{NestedProp:val2} or NotExisted:{Foo:Bar}",
		NewReflectEvaluator(obj),
		false)

	testExpr(
		t,
		"Prop:val and Nested:{notexisted:val2} or NotExisted:{Foo:Bar}",
		NewReflectEvaluator(obj),
		false)

	testExpr(
		t,
		"Nested:{notexisted:val2} or NotExisted:{Foo:Bar} or Prop:val",
		NewReflectEvaluator(obj),
		true)
}

func TestNested(t *testing.T) {
	obj := map[string]interface{}{
		"propStr": "value1",
		"nested": map[string]interface{}{
			"int": 13,
			"nested2": map[string]interface{}{
				"int": 42,
			},
		},
	}

	testExprMap(t, "propStr:'value2' or nested.int:13", obj, true)
	testExprMap(t, "propStr:'value2' or nested:{int:13}", obj, true)
	testExprMap(t, "propStr:'value2' or nested.nested2:{int:42}", obj, true)
	testExprMap(t, "propStr:'value2' or nested.nonexisted:{int:42}", obj, false)
	testExprMap(t, "propStr:'value2' or nonexisted.nonexisted:{int:42}", obj, false)
}

func TestTime(t *testing.T) {
	getTime := func(value string) time.Time {
		timeValue, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatal(err)
		}
		return timeValue
	}

	testType(
		t,
		"Ptime",
		"2021-05-17T01:00:00Z",
		"2021-05-18T01:00:00Z",
		"2021-05-16T01:00:00Z",
		testStruct{Ptime: getTime("2021-05-17T01:00:00Z")})
}

func TestDuration(t *testing.T) {
	getDuration := func(value string) time.Duration {
		duration, err := time.ParseDuration(value)
		if err != nil {
			t.Fatal(err)
		}
		return duration
	}

	testType(t, "Pduration", "300ms", "400ms", "200ms", testStruct{Pduration: getDuration("300ms")})
}

func TestString(t *testing.T) {
	testType(t, "Pstring", "2", "3", "1", testStruct{Pstring: "2"})
}

func TestInt64(t *testing.T) {
	testType(t, "Pint64", "2", "3", "1", testStruct{Pint64: 2})
}

func TestUint64(t *testing.T) {
	testType(t, "Puint64", "2", "3", "1", testStruct{Puint64: 2})
}

func TestFloat64(t *testing.T) {
	testType(t, "Pfloat64", "2.0", "3.0", "1.0", testStruct{Pfloat64: 2.0})
}

func TestTypeAliases(t *testing.T) {
	test(t, "Pint:1", true, testStruct{Pint: 1})
	test(t, "Pint8:1", true, testStruct{Pint8: 1})
	test(t, "Pint16:1", true, testStruct{Pint16: 1})
	test(t, "Pint32:1", true, testStruct{Pint32: 1})
	test(t, "Puint:1", true, testStruct{Puint: 1})
	test(t, "Puint8:1", true, testStruct{Puint8: 1})
	test(t, "Puint16:1", true, testStruct{Puint16: 1})
	test(t, "Puint32:1", true, testStruct{Puint32: 1})
	test(t, "Pfloat32:1", true, testStruct{Pfloat32: 1.0})
}

func testExprMap(t *testing.T, expression string, obj map[string]interface{}, expectedResult bool) {
	ev, err := NewMapEvaluator(obj)
	if err != nil {
		t.Fatal(err)
	}
	testExpr(t, expression, ev, expectedResult)
}

func test(t *testing.T, expression string, expectedResult bool, obj testStruct) {
	testExpr(t, expression, NewReflectEvaluator(obj), expectedResult)
}

func testType(t *testing.T, propertyName string, equalValue string, greaterValue string, lessValue string, obj testStruct) {
	testOp := func(op string, value string, expected bool) {
		test(t, propertyName+op+"'"+value+"'", expected, obj)
	}

	testOp(":", equalValue, true)
	testOp(":", greaterValue, false)
	testOp(":", lessValue, false)

	testOp(">", equalValue, false)
	testOp(">", greaterValue, false)
	testOp(">", lessValue, true)

	testOp(">=", equalValue, true)
	testOp(">=", greaterValue, false)
	testOp(">=", lessValue, true)

	testOp("<", equalValue, false)
	testOp("<", greaterValue, true)
	testOp("<", lessValue, false)

	testOp("<=", equalValue, true)
	testOp("<=", greaterValue, true)
	testOp("<=", lessValue, false)
}

func testExpr(t *testing.T, expression string, evaluator Evaluator, expectedResult bool) {
	t.Helper()
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
