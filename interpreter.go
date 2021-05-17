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
	return expression.ast.match(evaluator)
}

func (prop propertyMatch) match(evaluator Evaluator) (bool, error) {
	if prop.AtomicValue != nil {
		return matchAtomicValue(evaluator, prop)
	} else if prop.ValueSubExpression != nil {
		return matchSubExpression(evaluator, prop)
	} else if prop.OrValues != nil {
		return matchOrValues(evaluator, prop)
	} else if prop.AndValues != nil {
		return matchAndValues(evaluator, prop)
	}

	return false, errors.New("not implemented")
}

func matchAtomicValue(evaluator Evaluator, prop propertyMatch) (bool, error) {
	property, err := evaluator.Evaluate(prop.Name)
	if err != nil {
		return false, err
	}
	return equal(property, prop.AtomicValue)
}

func matchSubExpression(evaluator Evaluator, prop propertyMatch) (bool, error) {
	subEvaluator, err := evaluator.GetSubEvaluator(prop.Name)
	if err != nil {
		return false, err
	}
	return prop.ValueSubExpression.match(subEvaluator)
}

func matchOrValues(evaluator Evaluator, prop propertyMatch) (bool, error) {
	property, err := evaluator.Evaluate(prop.Name)
	if err != nil {
		return false, err
	}

	propertyValue := reflect.ValueOf(property)
	kind := propertyValue.Kind()
	if kind != reflect.String && kind != reflect.Int && kind != reflect.Slice {
		return false, errors.New("not implemented")
	}

	if kind == reflect.Slice {
		sliceLen := propertyValue.Len()
		for i := 0; i < sliceLen; i++ {
			sliceItem := propertyValue.Index(i).Interface()

			for _, item := range prop.OrValues {
				itemValue, err := equal(sliceItem, &item)
				if err != nil {
					return false, err
				}
				if itemValue {
					return true, nil
				}
			}
		}
	} else {
		for _, item := range prop.OrValues {
			itemValue, err := equal(property, &item)
			if err != nil {
				return false, err
			}
			if itemValue {
				return true, nil
			}
		}
	}

	return false, nil
}

func matchAndValues(evaluator Evaluator, prop propertyMatch) (bool, error) {
	property, err := evaluator.Evaluate(prop.Name)
	if err != nil {
		return false, err
	}

	propertyValue := reflect.ValueOf(property)
	kind := propertyValue.Kind()
	if kind != reflect.Slice {
		return false, errors.New("property " + prop.Name + " is expected to be a slice")
	}

	sliceLen := propertyValue.Len()

	for _, item := range prop.AndValues {
		itemFound := false
		for i := 0; i < sliceLen; i++ {
			sliceItem := propertyValue.Index(i).Interface()
			itemValue, err := equal(sliceItem, &item)
			if err != nil {
				return false, err
			}
			if itemValue {
				itemFound = true
				break
			}
		}

		if !itemFound {
			return false, nil
		}
	}

	return true, nil
}

func (se subExpression) match(evaluator Evaluator) (bool, error) {
	var seValue bool
	var err error
	if se.SubExpression != nil {
		seValue, err = se.SubExpression.match(evaluator)
		if err != nil {
			return false, err
		}
	} else {
		seValue, err = se.Value.match(evaluator)
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

func (c conjunction) match(evaluator Evaluator) (bool, error) {
	result, err := c.LeftValue.match(evaluator)
	if err != nil {
		return false, err
	}

	for _, right := range c.RightValues {
		if !result {
			return false, nil
		}

		var rightResult bool
		rightResult, err := right.match(evaluator)
		if err != nil {
			return false, err
		}

		result = result && rightResult
	}

	return result, nil
}

func (d disjunction) match(evaluator Evaluator) (bool, error) {
	result, err := d.LeftValue.match(evaluator)
	if err != nil {
		return false, err
	}

	for _, right := range d.RightValues {
		if result {
			return true, nil
		}

		var rightResult bool
		rightResult, err := right.match(evaluator)
		if err != nil {
			return false, err
		}

		result = result || rightResult
	}

	return result, nil
}

func (e expression) match(evaluator Evaluator) (bool, error) {
	return e.Expr.match(evaluator)
}

func equal(property interface{}, atomic *atomicValue) (bool, error) {
	switch v := property.(type) {
	case string:
		return atomic.wildcard.Match(v), nil
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
