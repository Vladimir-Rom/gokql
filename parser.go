package gokql

import (
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer/stateful"
)

type Expression struct {
	ast *expression
}

type atomicValue struct {
	Value          string `@Literal | @QuotedString | @DquotedString`
	convertedValue interface{}
	wildcard       wildcard
}

type propertyMatch struct {
	Name               string        `@Literal`
	Operation          string        `@(':' | '<' | '>' | '<=' | '>=')`
	ValueSubExpression *expression   `( ('{' @@ '}')`
	AtomicValue        *atomicValue  `| @@`
	OrValues           []atomicValue `| ('(' @@ ('or' @@)+')')`
	AndValues          []atomicValue `| ('(' @@ ('and' @@)+')'))`
}

type subExpression struct {
	IsInverted    bool           `@"not"?`
	SubExpression *expression    `('(' @@ ')'`
	Value         *propertyMatch `| @@)`
}

type conjunction struct {
	LeftValue   subExpression   `@@`
	RightValues []subExpression `('and' @@)*`
}

type disjunction struct {
	LeftValue   conjunction   `@@`
	RightValues []conjunction `('or' @@)*`
}

type expression struct {
	Expr disjunction `@@`
}

var (
	lexer, _ = stateful.NewSimple([]stateful.Rule{
		{"QuotedString", `'[^']*'`, nil},
		{"DquotedString", `"[^"]*"`, nil},
		{"Literal", `[a-zA-Z0-9.*]+`, nil},
		{"<=", `<=`, nil},
		{">=", `>=`, nil},
		{"whitespace", `[ \t\r\n]+`, nil},
		{"Any", ".", nil},
	})

	parser = participle.MustBuild(
		&expression{},
		participle.Lexer(lexer),
		participle.UseLookahead(10))
)

func parse(query string) (*expression, error) {
	var expr expression
	err := parser.ParseString(query, &expr)
	if err != nil {
		return nil, err
	}

	visitor := visitor{}
	visitor.atomicValue = func(atomic *atomicValue) {
		atomic.Value = unquote(atomic.Value)
		atomic.wildcard = newWildcard(atomic.Value)
	}
	expr.visit(visitor)

	return &expr, err
}

func Parse(query string) (*Expression, error) {
	ast, err := parse(query)
	if err != nil {
		return nil, err
	}

	return &Expression{ast}, nil
}

type visitor struct {
	atomicValue   func(*atomicValue)
	propertyMatch func(*propertyMatch)
	subExpression func(*subExpression)
	conjunction   func(*conjunction)
	disjunction   func(*disjunction)
	expression    func(*expression)
}

func (atomic atomicValue) String() string {
	return atomic.Value
}

func (prop propertyMatch) String() string {
	var valueStr string
	if prop.ValueSubExpression != nil {
		valueStr = "{" + prop.ValueSubExpression.String() + "}"
	} else if prop.AtomicValue != nil {
		valueStr = prop.AtomicValue.String()
	} else if prop.OrValues != nil {
		valueStr += "("
		for i, orValue := range prop.OrValues {
			if i == 0 {
				valueStr += orValue.String()
			} else {
				valueStr += " or " + orValue.String()
			}
		}
		valueStr += ")"
	} else if prop.AndValues != nil {
		valueStr += "("
		for i, orValue := range prop.OrValues {
			if i == 0 {
				valueStr += orValue.String()
			} else {
				valueStr += " and " + orValue.String()
			}
		}
		valueStr += ")"
	}

	return prop.Name + prop.Operation + valueStr
}

func (expr subExpression) String() string {
	notPrefix := ""
	if expr.IsInverted {
		notPrefix = "not "
	}

	if expr.SubExpression != nil {
		return notPrefix + expr.SubExpression.String()
	} else {
		return notPrefix + expr.Value.String()
	}
}

func (c conjunction) String() string {
	result := c.LeftValue.String()
	if c.RightValues != nil {
		for _, v := range c.RightValues {
			result += " and " + v.String()
		}

		result = "(" + result + ")"
	}

	return result
}

func (d disjunction) String() string {
	result := d.LeftValue.String()
	if d.RightValues != nil {
		for _, v := range d.RightValues {
			result += " or " + v.String()
		}

		result = "(" + result + ")"
	}

	return result
}

func (expr expression) String() string {
	return expr.Expr.String()
}

func unquote(str string) string {
	if len(str) == 0 {
		return str
	}

	if str[0] == '"' {
		return strings.Trim(str, "\"")
	}

	if str[0] == '\'' {
		return strings.Trim(str, "'")
	}

	return str
}

func (expr *expression) visit(visitor visitor) {
	expr.Expr.visit(visitor)
	if visitor.expression != nil {
		visitor.expression(expr)
	}
}

func (d *disjunction) visit(visitor visitor) {
	d.LeftValue.visit(visitor)
	for _, v := range d.RightValues {
		v.visit(visitor)
	}

	if visitor.disjunction != nil {
		visitor.disjunction(d)
	}
}

func (c *conjunction) visit(visitor visitor) {
	c.LeftValue.visit(visitor)
	for _, v := range c.RightValues {
		v.visit(visitor)
	}

	if visitor.conjunction != nil {
		visitor.conjunction(c)
	}
}

func (e *subExpression) visit(visitor visitor) {
	if e.SubExpression != nil {
		e.SubExpression.visit(visitor)
	}
	if e.Value != nil {
		e.Value.visit(visitor)
	}

	if visitor.subExpression != nil {
		visitor.subExpression(e)
	}
}

func (pm *propertyMatch) visit(visitor visitor) {
	if pm.AtomicValue != nil {
		pm.AtomicValue.visit(visitor)
	}
	if pm.ValueSubExpression != nil {
		pm.ValueSubExpression.visit(visitor)
	}
	if pm.OrValues != nil {
		for _, orValue := range pm.OrValues {
			orValue.visit(visitor)
		}
	}
	if pm.AndValues != nil {
		for _, andValue := range pm.AndValues {
			andValue.visit(visitor)
		}
	}

	if visitor.propertyMatch != nil {
		visitor.propertyMatch(pm)
	}
}

func (atomic *atomicValue) visit(visitor visitor) {
	if visitor.atomicValue != nil {
		visitor.atomicValue(atomic)
	}
}
