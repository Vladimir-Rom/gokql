package gokql

import (
	"fmt"
	"reflect"
)

type EvaluatorKind string

const (
	EvaluatorKindObject EvaluatorKind = "object"
	EvaluatorKindSlice  EvaluatorKind = "array"
)

type Evaluator interface {
	Evaluate(propertyName string) (interface{}, error)
	GetSubEvaluator(propertyName string) (Evaluator, error)
	GetEvaluatorKind() EvaluatorKind
	GetArraySubEvaluators() ([]Evaluator, error)
}

type NullEvaluator struct {
}

func (NullEvaluator) Evaluate(propertyName string) (interface{}, error) {
	return nil, nil
}

func (NullEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	return NullEvaluator{}, nil
}

func (NullEvaluator) GetEvaluatorKind() EvaluatorKind {
	return EvaluatorKindObject
}

func (NullEvaluator) GetArraySubEvaluators() ([]Evaluator, error) {
	return nil, nil
}

type MapEvaluator struct {
	obj  map[string]any
	arr  []map[string]any
	kind EvaluatorKind
}

func NewMapEvaluator(item any) (*MapEvaluator, error) {
	switch v := any(item).(type) {
	case map[string]any:
		return &MapEvaluator{
			obj:  v,
			kind: EvaluatorKindObject,
		}, nil
	case []map[string]any:
		return &MapEvaluator{
			arr:  v,
			kind: EvaluatorKindSlice,
		}, nil
	case []any:
		arr := make([]map[string]any, len(v))
		for i := range arr {
			if arrItem, ok := v[i].(map[string]any); ok {
				arr[i] = arrItem
			} else {
				return nil, fmt.Errorf("unexpected array item type: %T. It should be map[string]any", v[i])
			}
		}
		return &MapEvaluator{
			arr:  arr,
			kind: EvaluatorKindSlice,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected item type: %T", item)
	}
}

func (m *MapEvaluator) Evaluate(propertyName string) (interface{}, error) {
	if result, ok := m.obj[propertyName]; !ok {
		return nil, nil
	} else {
		return result, nil
	}
}

func (m *MapEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
	if result, ok := m.obj[propertyName]; !ok {
		return nil, nil
	} else {
		return NewMapEvaluator(result)
	}
}

func (m *MapEvaluator) GetEvaluatorKind() EvaluatorKind {
	return m.kind
}

func (m *MapEvaluator) GetArraySubEvaluators() ([]Evaluator, error) {
	if m.GetEvaluatorKind() != EvaluatorKindSlice {
		return nil, fmt.Errorf("unsupported operation for kind %v", m.kind)
	}

	res := make([]Evaluator, len(m.arr))
	for i := range res {
		ev, err := NewMapEvaluator(m.arr[i])
		if err != nil {
			return nil, err
		}
		res[i] = ev
	}
	return res, nil
}

type ReflectEvaluator struct {
	value *reflect.Value
}

func NewReflectEvaluator(value interface{}) *ReflectEvaluator {
	val := reflect.ValueOf(value)
	return &ReflectEvaluator{&val}
}

func (eval *ReflectEvaluator) Evaluate(propertyName string) (interface{}, error) {
	if eval.value == nil {
		panic("evaluator is not initialized")
	}

	res := eval.value.FieldByName(propertyName)
	if res == (reflect.Value{}) {
		return nil, nil
	}

	return res.Interface(), nil
}

func (eval *ReflectEvaluator) GetSubEvaluator(propertyName string) (Evaluator, error) {
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

func (eval *ReflectEvaluator) GetEvaluatorKind() EvaluatorKind {
	if eval.value.Kind() == reflect.Slice {
		return EvaluatorKindSlice
	}
	return EvaluatorKindObject
}

func (eval *ReflectEvaluator) GetArraySubEvaluators() ([]Evaluator, error) {
	if eval.GetEvaluatorKind() != EvaluatorKindSlice {
		return nil, fmt.Errorf("unsupported operation for kind %v", eval.GetEvaluatorKind())
	}
	res := make([]Evaluator, eval.value.Len())
	for i := range res {
		res[i] = NewReflectEvaluator(eval.value.Index(i).Interface())
	}
	return res, nil
}
