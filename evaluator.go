package gokql

import (
	"errors"
	"reflect"
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

type ReflectEvaluator struct {
	value reflect.Value
}

func NewReflectEvaluator(value interface{}) ReflectEvaluator {
	return ReflectEvaluator{reflect.ValueOf(value)}
}

func (eval ReflectEvaluator) Evaluate(propertyName string) (interface{}, error) {
	res := eval.value.FieldByName(propertyName)
	if res == (reflect.Value{}) {
		return nil, errors.New("property " + propertyName + " not found")
	}

	return res.Interface(), nil
}

func (eval ReflectEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	prop, err := eval.Evaluate(propertyName)
	if err != nil {
		return nil, err
	}

	return NewReflectEvaluator(prop), nil
}
