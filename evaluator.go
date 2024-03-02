package gokql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Evaluator interface {
	Evaluate(propertyName string) (interface{}, error)
	GetSubEvaluator(propertyName string) (Evaluator, error)
}

type NestedPropsEvaluator struct {
	inner Evaluator
}

func (e NestedPropsEvaluator) drilldown(propertyName string) (ev Evaluator, property string, err error) {
	ev = e.inner
	parts := strings.Split(propertyName, ".")
	for i, part := range parts[:len(parts)-1] {
		var err error
		evPrev := ev
		ev, err = ev.GetSubEvaluator(part)
		if err != nil {
			return nil, "", fmt.Errorf("unable to find property %s, Error: %w", part, err)
		}
		if ev == nil {
			return evPrev, strings.Join(parts[i:], "."), nil
		}
	}
	return ev, parts[len(parts)-1], nil
}

func (e NestedPropsEvaluator) Evaluate(propertyName string) (interface{}, error) {
	ev, prop, err := e.drilldown(propertyName)
	if err != nil {
		return nil, err
	}

	return ev.Evaluate(prop)
}

func (e NestedPropsEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	ev, prop, err := e.drilldown(propertyName)
	if err != nil {
		return nil, err
	}

	return ev.GetSubEvaluator(prop)
}

type NullEvaluator struct {
}

func (NullEvaluator) Evaluate(propertyName string) (interface{}, error) {
	return nil, nil
}

func (NullEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	return NullEvaluator{}, nil
}

type MapEvaluator struct {
	Map map[string]interface{}
}

func (m MapEvaluator) Evaluate(propertyName string) (interface{}, error) {
	if result, ok := m.Map[propertyName]; !ok {
		return nil, nil
	} else {
		return result, nil
	}
}

func (m MapEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	if result, ok := m.Map[propertyName]; !ok {
		return nil, nil
	} else {
		if innerMap, ok := result.(map[string]interface{}); !ok {
			return nil, errors.New("property " + propertyName + " expected to be a 'map[string] interface{}'")
		} else {
			return MapEvaluator{innerMap}, nil
		}
	}
}

type ReflectEvaluator struct {
	value *reflect.Value
}

func NewReflectEvaluator(value interface{}) ReflectEvaluator {
	val := reflect.ValueOf(value)
	return ReflectEvaluator{&val}
}

func (eval ReflectEvaluator) Evaluate(propertyName string) (interface{}, error) {
	if eval.value == nil {
		panic("evaluator is not initialized")
	}

	res := eval.value.FieldByName(propertyName)
	if res == (reflect.Value{}) {
		return nil, nil
	}

	return res.Interface(), nil
}

func (eval ReflectEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	if eval.value == nil {
		panic("evaluator is not initialized")
	}

	prop, err := eval.Evaluate(propertyName)
	if err != nil {
		return nil, err
	}

	if prop == nil {
		return NullEvaluator{}, nil
	}

	return NewReflectEvaluator(prop), nil
}
