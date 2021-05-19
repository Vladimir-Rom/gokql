# gokql - Kibana Query Language (KQL) interpreter for Go

Kibana Query Language (KQL) - it is a simple query language for filtering objects used in Kibana. KQL syntax is described here: https://www.elastic.co/guide/en/kibana/master/kuery-query.html

gokql - is a Go package for embedding KQL into Go applications. It can be used in a command line utility to receive filters from a command line. Example of such utility is the docker tool:  
```shell
$ docker ps --filter "label=value"
```

If docker used KQL for filters, the command would be as follows:  
```shell
$ docker ps --filter "label:value"
```

To filer your data using gokql you need parse query by method goqkl.Parse, then call method Match on returned value:

```go
// Parse query
expression, err := gokql.Parse("Label:value1")
...
// then perform matching
someItem := struct {
    Label []string {"value1", "value2"}
}
matched, err := expression.Match(gokql.NewReflectEvaluator(someItem))
if err != nil {
    ...
} else {
    fmt.Printf("Matched: %v", matched)
}
```

Method Match accepts interface gokql.Evaluator:
```go
type Evaluator interface {
	Evaluate(propertyName string) (interface{}, error)
	GetSubEvaluator(propertyName string) (Evaluator, error)
}
```

Interface `Evaluator` is used for evaluating property values when matching them with query. Gokql provides two implementations of the Evaluator interface: `MapEvaluator` for filtering over maps and presented above `ReflectEvaluator` for filtering over structs using reflection. Example for filtering over maps:

```go
// Parse query
expression, err := gokql.Parse("Label:value1")
...
// then perform matching
someItem := map[string]interface{} {
    "Label": []string {"value1", "value2"},
}
matched, err := expression.Match(MapEvaluator{someItem})
```

Method Evaluator.GetSubEvaluator is used by qokql for nested query syntax: `nestedProperty:{value:foo}`:

```go
// Parse query
expression, err := gokql.Parse("nestedProperty:{value:foo}")
...
// then perform matching
someItem := map[string]interface{} {
    "nestedProperty": map[string]interface{} {
        "value": "foo",
    },
}
matched, err := expression.Match(MapEvaluator{someItem})
```


For performance reasons don't parse queries for each data item. It is better to parse a query once, save parsed expression and then use it over collection of filtering objects. Parsed expression is thread safe and can be used in different goroutines. 