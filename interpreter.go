package gokql

import (
	"errors"
	"reflect"
	"strconv"
)

type Evaluator interface {
	Evaluate(propertyName string) (interface{}, error)
	GetSubEvaluator(propertyName string) (Evaluator, error)
}

type MapEvaluator struct {
	Map map[string]interface{}
}

func (m MapEvaluator) Evaluate(propertyName string) (interface{}, error) {
	if result, ok := m.Map[propertyName]; !ok {
		return nil, errors.New("Property " + propertyName + " not found")
	} else {
		return result, nil
	}
}

func (m MapEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	if result, ok := m.Map[propertyName]; !ok {
		return nil, errors.New("property " + propertyName + " not found")
	} else {
		if innerMap, ok := result.(map[string]interface{}); !ok {
			return nil, errors.New("property " + propertyName + " expected to be a 'map[string] interface{}'")
		} else {
			return MapEvaluator{innerMap}, nil
		}
	}
}

func (expression Expression) Match(evaluator Evaluator) (bool, error) {
	return expression.ast.Match(evaluator)
}

func (prop propertyMatch) Match(evaluator Evaluator) (bool, error) {
	if prop.AtomicValue != nil {
		property, err := evaluator.Evaluate(prop.Name)
		if err != nil {
			return false, err
		}
		return equal(property, prop.AtomicValue)
	} else if prop.ValueSubExpression != nil {
		subEvaluator, err := evaluator.GetSubEvaluator(prop.Name)
		if err != nil {
			return false, err
		}
		return prop.ValueSubExpression.Match(subEvaluator)
	} else if prop.OrValues != nil {
		property, err := evaluator.Evaluate(prop.Name)
		if err != nil {
			return false, err
		}

		propertyValue := reflect.ValueOf(property)
		kind := propertyValue.Kind()
		if kind != reflect.String && kind != reflect.Int {
			return false, errors.New("not implemented")
		}

		result := false
		for _, item := range prop.OrValues {
			if result {
				return true, nil
			}
			itemValue, err := equal(property, &item)
			if err != nil {
				return false, err
			}
			result = result || itemValue
		}
		return result, nil
	}

	return false, errors.New("not implemented")
}

func (se subExpression) Match(evaluator Evaluator) (bool, error) {
	var seValue bool
	var err error
	if se.SubExpression != nil {
		seValue, err = se.SubExpression.Match(evaluator)
		if err != nil {
			return false, err
		}
	} else {
		seValue, err = se.Value.Match(evaluator)
		if err != nil {
			return false, err
		}
	}

	if se.IsInverted {
		return !seValue, nil
	} else {
		return seValue, nil
	}
}

func (c conjunction) Match(evaluator Evaluator) (bool, error) {
	result, err := c.LeftValue.Match(evaluator)
	if err != nil {
		return false, err
	}

	for _, right := range c.RightValues {
		if !result {
			return false, nil
		}

		var rightResult bool
		rightResult, err := right.Match(evaluator)
		if err != nil {
			return false, err
		}

		result = result && rightResult
	}

	return result, nil
}

func (d disjunction) Match(evaluator Evaluator) (bool, error) {
	result, err := d.LeftValue.Match(evaluator)
	if err != nil {
		return false, err
	}

	for _, right := range d.RightValues {
		if result {
			return true, nil
		}

		var rightResult bool
		rightResult, err := right.Match(evaluator)
		if err != nil {
			return false, err
		}

		result = result || rightResult
	}

	return result, nil
}

func (e expression) Match(evaluator Evaluator) (bool, error) {
	return e.Expr.Match(evaluator)
}

func equal(property interface{}, atomic *atomicValue) (bool, error) {
	switch v := property.(type) {
	case string:
		return v == atomic.Value, nil
	case int:
		convertedValue := atomic.convertedValue
		if intValue, ok := convertedValue.(int); ok {
			return v == intValue, nil
		}
		intValue, err := strconv.Atoi(atomic.Value)
		if err != nil {
			return false, err
		}
		atomic.convertedValue = intValue
		return v == intValue, nil
	default:
		return false, errors.New("Unsupported property type: " + reflect.TypeOf(property).Name())
	}
}
