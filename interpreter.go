package gokql

import (
	"errors"
	"reflect"
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
			res, err := compare(propertyValue.Index(i).Interface(), prop.AtomicValue, equalComparer{})
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
			return compare(property, prop.AtomicValue, equalComparer{})
		case ">":
			return compare(property, prop.AtomicValue, greaterComparer{})
		case ">=":
			return compare(property, prop.AtomicValue, greaterOrEqualComparer{})
		case "<":
			return compare(property, prop.AtomicValue, lessComparer{})
		case "<=":
			return compare(property, prop.AtomicValue, lessOrEqualComparer{})
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
				itemValue, err := compare(sliceItem, &item, equalComparer{})
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
			itemValue, err := compare(property, &item, equalComparer{})
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
			itemValue, err := compare(sliceItem, &item, equalComparer{})
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

func compare(property interface{}, atomic *atomicValue, comparer comparer) (bool, error) {
	switch v := property.(type) {
	case string:
		left := v
		right := atomic.Value
		res, _ := comparer.compare(
			compareOperations{
				equal:          func() (bool, bool) { return atomic.wildcard.Match(left), true },
				greater:        func() (bool, bool) { return left > right, true },
				less:           func() (bool, bool) { return left < right, true },
				greaterOrEqual: func() (bool, bool) { return left >= right, true },
				lessOrEqual:    func() (bool, bool) { return left <= right, true },
			})
		return res, nil

	case int:
		return compareValues(atomic, comparer, int64TypeHandler{}, int64(v))
	case int8:
		return compareValues(atomic, comparer, int64TypeHandler{}, int64(v))
	case int16:
		return compareValues(atomic, comparer, int64TypeHandler{}, int64(v))
	case int32:
		return compareValues(atomic, comparer, int64TypeHandler{}, int64(v))
	case int64:
		return compareValues(atomic, comparer, int64TypeHandler{}, v)
	case uint:
		return compareValues(atomic, comparer, uint64TypeHandler{}, uint64(v))
	case uint8:
		return compareValues(atomic, comparer, uint64TypeHandler{}, uint64(v))
	case uint16:
		return compareValues(atomic, comparer, uint64TypeHandler{}, uint64(v))
	case uint32:
		return compareValues(atomic, comparer, uint64TypeHandler{}, uint64(v))
	case uint64:
		return compareValues(atomic, comparer, uint64TypeHandler{}, v)
	case bool:
		return compareValues(atomic, comparer, boolTypeHandler{}, v)
	case float64:
		return compareValues(atomic, comparer, float64TypeHandler{}, v)
	case float32:
		return compareValues(atomic, comparer, float64TypeHandler{}, float64(v))
	case time.Time:
		return compareValues(atomic, comparer, timeTypeHandler{}, v)
	case time.Duration:
		return compareValues(atomic, comparer, durationTypeHandler{}, v)

	default:
		return false, errors.New("Unsupported property type: " + reflect.TypeOf(property).Name())
	}
}

type comparer interface {
	compare(compareOperations) (result bool, ok bool)
}

type compareOperations struct {
	equal          func() (result bool, ok bool)
	greater        func() (result bool, ok bool)
	less           func() (result bool, ok bool)
	greaterOrEqual func() (result bool, ok bool)
	lessOrEqual    func() (result bool, ok bool)
}

type equalComparer struct{}

func (equalComparer) compare(compareOperations compareOperations) (result bool, ok bool) {
	return compareOperations.equal()
}

type greaterComparer struct{}

func (greaterComparer) compare(compareOperations compareOperations) (result bool, ok bool) {
	return compareOperations.greater()
}

type greaterOrEqualComparer struct{}

func (greaterOrEqualComparer) compare(compareOperations compareOperations) (result bool, ok bool) {
	return compareOperations.greaterOrEqual()
}

type lessComparer struct{}

func (lessComparer) compare(compareOperations compareOperations) (result bool, ok bool) {
	return compareOperations.less()
}

type lessOrEqualComparer struct{}

func (lessOrEqualComparer) compare(compareOperations compareOperations) (result bool, ok bool) {
	return compareOperations.lessOrEqual()
}

func compareValues(atomic *atomicValue, comparer comparer, handler typeHandler, value interface{}) (bool, error) {
	compare := func(left interface{}, right interface{}) (result bool, ok bool) {
		return comparer.compare(
			compareOperations{
				equal:          func() (bool, bool) { return handler.equal(left, right) },
				greater:        func() (bool, bool) { return handler.greater(left, right) },
				less:           func() (bool, bool) { return handler.less(left, right) },
				greaterOrEqual: func() (bool, bool) { return handler.greaterOrEqual(left, right) },
				lessOrEqual:    func() (bool, bool) { return handler.lessOrEqual(left, right) },
			})
	}

	convertedValue := atomic.convertedValue
	if res, ok := compare(convertedValue, value); ok {
		return res, nil
	}

	convertedValue, err := handler.convert(atomic.Value)
	if err != nil {
		return false, err
	}

	atomic.convertedValue = convertedValue
	res, _ := compare(convertedValue, value)
	return res, nil
}
