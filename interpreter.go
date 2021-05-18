package gokql

import (
	"errors"
	"reflect"
	"strconv"
	"time"
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

	propertyValue := reflect.ValueOf(property)

	if propertyValue.Kind() == reflect.Slice {
		sliceLen := propertyValue.Len()
		for i := 0; i < sliceLen; i++ {
			res, err := compare(propertyValue.Index(i).Interface(), prop.AtomicValue, equalOp{})
			if err != nil {
				return false, err
			}
			if res {
				return true, nil
			}
		}
		return false, nil
	} else {
		switch prop.Operation {
		case ":":
			return compare(property, prop.AtomicValue, equalOp{})
		case ">":
			return compare(property, prop.AtomicValue, greaterOp{})
		case ">=":
			return compare(property, prop.AtomicValue, greaterOrEqualOp{})
		case "<":
			return compare(property, prop.AtomicValue, lessOp{})
		case "<=":
			return compare(property, prop.AtomicValue, lessOrEqualOp{})
		default:
			panic("unknown operation " + prop.Operation)
		}
	}

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
				itemValue, err := compare(sliceItem, &item, equalOp{})
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
			itemValue, err := compare(property, &item, equalOp{})
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
			itemValue, err := compare(sliceItem, &item, equalOp{})
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

type operation interface {
	compareStr(string, string, wildcard) bool
	compareInt(int64, int64) bool
	compareTime(time.Time, time.Time) bool
}

type equalOp struct{}

func (equalOp) compareStr(left string, right string, wildcard wildcard) bool {
	return wildcard.Match(left)
}

func (equalOp) compareInt(left int64, right int64) bool {
	return left == right
}

func (equalOp) compareUint(left uint64, right uint64) bool {
	return left == right
}

func (equalOp) compareTime(left time.Time, right time.Time) bool {
	return left.Equal(right)
}

type greaterOp struct{}

func (greaterOp) compareStr(left string, right string, wildcard wildcard) bool {
	return left > right
}

func (greaterOp) compareInt(left int64, right int64) bool {
	return left > right
}

func (greaterOp) compareUint(left uint64, right uint64) bool {
	return left > right
}

func (greaterOp) compareTime(left time.Time, right time.Time) bool {
	return left.After(right)
}

type greaterOrEqualOp struct{}

func (greaterOrEqualOp) compareStr(left string, right string, wildcard wildcard) bool {
	return left >= right
}

func (greaterOrEqualOp) compareInt(left int64, right int64) bool {
	return left >= right
}

func (greaterOrEqualOp) compareTime(left time.Time, right time.Time) bool {
	return left.Equal(right) || left.After(right)
}

type lessOp struct{}

func (lessOp) compareStr(left string, right string, wildcard wildcard) bool {
	return left < right
}

func (lessOp) compareInt(left int64, right int64) bool {
	return left < right
}

func (lessOp) compareTime(left time.Time, right time.Time) bool {
	return left.Before(right)
}

type lessOrEqualOp struct{}

func (lessOrEqualOp) compareStr(left string, right string, wildcard wildcard) bool {
	return left <= right
}

func (lessOrEqualOp) compareInt(left int64, right int64) bool {
	return left <= right
}

func (lessOrEqualOp) compareTime(left time.Time, right time.Time) bool {
	return left.Equal(right) || left.Before(right)
}

func compare(property interface{}, atomic *atomicValue, operation operation) (bool, error) {
	switch v := property.(type) {
	case string:
		return operation.compareStr(v, atomic.Value, atomic.wildcard), nil
	case byte:
		return compareInt(atomic, operation, int64(v))
	case int:
		return compareInt(atomic, operation, int64(v))
	case int8:
		return compareInt(atomic, operation, int64(v))
	case int16:
		return compareInt(atomic, operation, int64(v))
	case int32:
		return compareInt(atomic, operation, int64(v))
	case int64:
		return compareInt(atomic, operation, v)
	case time.Time:
		convertedValue := atomic.convertedValue
		if timeValue, ok := convertedValue.(time.Time); ok {
			return operation.compareTime(v, timeValue), nil
		}
		timeValue, err := time.Parse(time.RFC3339, atomic.Value)
		if err != nil {
			return false, err
		}
		atomic.convertedValue = timeValue
		return operation.compareTime(v, timeValue), nil
	default:
		return false, errors.New("Unsupported property type: " + reflect.TypeOf(property).Name())
	}
}

func compareInt(atomic *atomicValue, operation operation, v int64) (bool, error) {
	convertedValue := atomic.convertedValue
	if intValue, ok := convertedValue.(int64); ok {
		return operation.compareInt(v, intValue), nil
	}
	intValue, err := strconv.ParseInt(atomic.Value, 10, 64)
	if err != nil {
		return false, err
	}
	atomic.convertedValue = intValue
	return operation.compareInt(v, intValue), nil
}
