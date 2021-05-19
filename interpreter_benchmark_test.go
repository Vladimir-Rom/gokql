package gokql

import (
	"testing"
)

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
