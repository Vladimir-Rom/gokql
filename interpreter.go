package gokql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

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
	property, err := evaluateWithDrilldown(evaluator, prop.Name)
	if err != nil {
		return false, err
	}

	if property == nil {
		return false, nil
	}

	propertyValue := reflect.ValueOf(property)

	if propertyValue.Kind() == reflect.Slice {
		sliceLen := propertyValue.Len()
		for i := 0; i < sliceLen; i++ {
			res, err := compare(propertyValue.Index(i).Interface(), prop.AtomicValue, equalCmp{})
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
			return compare(property, prop.AtomicValue, equalCmp{})
		case ">":
			return compare(property, prop.AtomicValue, greaterCmp{})
		case ">=":
			return compare(property, prop.AtomicValue, greaterOrEqualCmp{})
		case "<":
			return compare(property, prop.AtomicValue, lessCmp{})
		case "<=":
			return compare(property, prop.AtomicValue, lessOrEqualCmp{})
		default:
			panic("unknown operation " + prop.Operation)
		}
	}

}

func matchSubExpression(evaluator Evaluator, prop propertyMatch) (bool, error) {
	subEvaluator, err := drilldownEvaluator(prop.Name, evaluator)
	if err != nil {
		return false, err
	}

	if subEvaluator == nil {
		return false, nil
	}

	if subEvaluator.GetEvaluatorKind() == EvaluatorKindObject {
		return prop.ValueSubExpression.match(subEvaluator)
	}

	sliceEvals, err := subEvaluator.GetArraySubEvaluators()
	if err != nil {
		return false, err
	}

	for _, ev := range sliceEvals {
		res, err := prop.ValueSubExpression.match(ev)
		if err != nil {
			return false, err
		}

		if res {
			return true, nil
		}
	}
	return false, nil
}

func drilldownEvaluator(propertyNames []string, ev Evaluator) (Evaluator, error) {
	for _, name := range propertyNames {
		subev, err := ev.GetSubEvaluator(name)
		if err != nil {
			return nil, fmt.Errorf("unable to find property %s, Error: %w", name, err)
		}
		if subev == nil {
			return nil, nil
		}
		ev = subev
	}
	return ev, nil
}

func evaluateWithDrilldown(evaluator Evaluator, propName []string) (any, error) {
	subEvaluator, err := drilldownEvaluator(propName[:len(propName)-1], evaluator)
	if err != nil {
		return false, err
	}

	if subEvaluator == nil {
		return false, nil
	}

	return subEvaluator.Evaluate(propName[len(propName)-1])
}

func matchOrValues(evaluator Evaluator, prop propertyMatch) (bool, error) {
	property, err := evaluateWithDrilldown(evaluator, prop.Name)
	if err != nil {
		return false, err
	}

	if property == nil {
		return false, nil
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
				itemValue, err := compare(sliceItem, &item, equalCmp{})
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
			itemValue, err := compare(property, &item, equalCmp{})
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
	property, err := evaluateWithDrilldown(evaluator, prop.Name)
	if err != nil {
		return false, err
	}
	if property == nil {
		return false, nil
	}

	propertyValue := reflect.ValueOf(property)
	kind := propertyValue.Kind()
	if kind != reflect.Slice {
		return false, errors.New("property " + strings.Join(prop.Name, ".") + " is expected to be a slice")
	}

	sliceLen := propertyValue.Len()

	for _, item := range prop.AndValues {
		itemFound := false
		for i := 0; i < sliceLen; i++ {
			sliceItem := propertyValue.Index(i).Interface()
			itemValue, err := compare(sliceItem, &item, equalCmp{})
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

type comparer interface {
	compare(handler typeHandler, left interface{}, right interface{}) bool
}

type equalCmp struct{}

func (equalCmp) compare(handler typeHandler, left interface{}, right interface{}) bool {
	return handler.equal(left, right)
}

type greaterCmp struct{}

func (greaterCmp) compare(handler typeHandler, left interface{}, right interface{}) bool {
	return handler.greater(left, right)
}

type lessCmp struct{}

func (lessCmp) compare(handler typeHandler, left interface{}, right interface{}) bool {
	return handler.less(left, right)
}

type greaterOrEqualCmp struct{}

func (greaterOrEqualCmp) compare(handler typeHandler, left interface{}, right interface{}) bool {
	return handler.greaterOrEqual(left, right)
}

type lessOrEqualCmp struct{}

func (lessOrEqualCmp) compare(handler typeHandler, left interface{}, right interface{}) bool {
	return handler.lessOrEqual(left, right)
}

func compare(property interface{}, atomic *atomicValue, comparer comparer) (bool, error) {
	switch v := property.(type) {
	case int:
		return compareWithConvertedType(int64(v), atomic, comparer)
	case int8:
		return compareWithConvertedType(int64(v), atomic, comparer)
	case int16:
		return compareWithConvertedType(int64(v), atomic, comparer)
	case int32:
		return compareWithConvertedType(int64(v), atomic, comparer)
	case uint:
		return compareWithConvertedType(uint64(v), atomic, comparer)
	case uint8:
		return compareWithConvertedType(uint64(v), atomic, comparer)
	case uint16:
		return compareWithConvertedType(uint64(v), atomic, comparer)
	case uint32:
		return compareWithConvertedType(uint64(v), atomic, comparer)
	case float32:
		return compareWithConvertedType(float64(v), atomic, comparer)

	default:
		return compareWithConvertedType(property, atomic, comparer)
	}
}

func compareWithConvertedType(property interface{}, atomic *atomicValue, comparer comparer) (bool, error) {
	if atomic.wildcard.firstStar && atomic.wildcard.lastStar && len(atomic.wildcard.parts) == 0 {
		return true, nil
	}

	propertyType := reflect.TypeOf(property)
	if atomic.comparer == nil || atomic.valueType != propertyType {
		var err error
		atomic.comparer, err = createComparer(property, atomic, comparer)
		if err != nil {
			return false, err
		}

		atomic.valueType = propertyType
	}

	return atomic.comparer(property), nil
}

func createComparer(propertyValue interface{}, atomic *atomicValue, comparer comparer) (func(propertyValue interface{}) bool, error) {
	switch propertyValue.(type) {
	case string:
		return createComparerForHandler(stringTypeHandler{atomic.wildcard}, propertyValue, atomic, comparer)
	case int64:
		return createComparerForHandler(int64TypeHandler{}, propertyValue, atomic, comparer)
	case uint64:
		return createComparerForHandler(uint64TypeHandler{}, propertyValue, atomic, comparer)
	case float64:
		return createComparerForHandler(float64TypeHandler{}, propertyValue, atomic, comparer)
	case bool:
		return createComparerForHandler(boolTypeHandler{}, propertyValue, atomic, comparer)
	case time.Time:
		return createComparerForHandler(timeTypeHandler{}, propertyValue, atomic, comparer)
	case time.Duration:
		return createComparerForHandler(durationTypeHandler{}, propertyValue, atomic, comparer)
	}

	return nil, errors.New("unsupported property type " + reflect.TypeOf(propertyValue).Name())
}

func createComparerForHandler(
	handler typeHandler,
	propertyValue interface{},
	atomic *atomicValue,
	comparer comparer) (func(propertyValue interface{}) bool, error) {
	requestValue, err := handler.convert(atomic.Value)
	if err != nil {
		return nil, err
	}

	return func(propertyValue interface{}) bool {
		return comparer.compare(handler, propertyValue, requestValue)
	}, nil
}
