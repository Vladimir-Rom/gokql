package gokql

import (
	"testing"
)

var expr Expression
var evaluator Evaluator

func init() {
	var err error
	expr, err = Parse("prop1: 42")
	if err != nil {
		panic(err)
	}

	evaluator, err = NewMapEvaluator(
		map[string]interface{}{
			"prop1":  42,
			"prop2:": "asdasdasd",
			"msg":    "asdasdac LDS CA, LDS CSD LSD ASD CSDqefosmd,md,mdklf",
			"level":  "info",
		})

	expr.Match(evaluator)
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expr.Match(evaluator)
	}
}
