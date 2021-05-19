package gokql

import (
	"strconv"
	"time"
)

type typeHandler interface {
	convert(value string) (result interface{}, err error)
	equal(left interface{}, right interface{}) bool
	greater(left interface{}, right interface{}) bool
	less(left interface{}, right interface{}) bool
	greaterOrEqual(left interface{}, right interface{}) bool
	lessOrEqual(left interface{}, right interface{}) bool
}

// ================ STRING  ====================}

type stringTypeHandler struct {
	wildcard wildcard
}

func (stringTypeHandler) convert(value string) (result interface{}, err error) {
	return value, nil
}

func (s stringTypeHandler) equal(left interface{}, right interface{}) bool {
	return s.wildcard.Match(left.(string))
}

func (s stringTypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(string) > right.(string)
}

func (s stringTypeHandler) less(left interface{}, right interface{}) bool {
	return left.(string) < right.(string)
}

func (s stringTypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	return left.(string) >= right.(string)
}

func (s stringTypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	return left.(string) <= right.(string)
}

// ================ INT64  ====================

type int64TypeHandler struct{}

func (int64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseInt(value, 10, 64)
}

func (i int64TypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(int64) == right.(int64)
}

func (i int64TypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(int64) > right.(int64)
}

func (i int64TypeHandler) less(left interface{}, right interface{}) bool {
	return left.(int64) < right.(int64)
}

func (i int64TypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	return left.(int64) >= right.(int64)
}

func (i int64TypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	return left.(int64) <= right.(int64)
}

// ================ UINT64  ====================

type uint64TypeHandler struct{}

func (uint64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseUint(value, 10, 64)
}

func (i uint64TypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(uint64) == right.(uint64)
}

func (i uint64TypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(uint64) > right.(uint64)
}

func (i uint64TypeHandler) less(left interface{}, right interface{}) bool {
	return left.(uint64) < right.(uint64)
}

func (i uint64TypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	return left.(uint64) >= right.(uint64)
}

func (i uint64TypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	return left.(uint64) <= right.(uint64)
}

// ================ TIME  ====================

type timeTypeHandler struct{}

func (timeTypeHandler) convert(value string) (result interface{}, err error) {
	return time.Parse(time.RFC3339, value)
}

func (t timeTypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(time.Time).Equal(right.(time.Time))
}

func (t timeTypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(time.Time).After(right.(time.Time))
}

func (t timeTypeHandler) less(left interface{}, right interface{}) bool {
	return left.(time.Time).Before(right.(time.Time))
}

func (t timeTypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	l := left.(time.Time)
	r := right.(time.Time)
	return l.After(r) || l.Equal(r)
}

func (t timeTypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	l := left.(time.Time)
	r := right.(time.Time)

	return l.Before(r) || l.Equal(r)
}

// ================ BOOL  ====================

type boolTypeHandler struct{}

func (boolTypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseBool(value)
}

func (t boolTypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(bool) == right.(bool)
}

func (t boolTypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(bool) && !right.(bool)
}

func (t boolTypeHandler) less(left interface{}, right interface{}) bool {
	return !left.(bool) && right.(bool)
}

func (t boolTypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	l := left.(bool)
	r := right.(bool)
	return (l && !r) || (l == r)
}

func (t boolTypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	l := left.(bool)
	r := right.(bool)
	return (!l && r) || (l == r)
}

// ================ FLOAT ====================

type float64TypeHandler struct{}

func (float64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseFloat(value, 64)
}

func (t float64TypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(float64) == right.(float64)
}

func (t float64TypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(float64) > right.(float64)
}

func (t float64TypeHandler) less(left interface{}, right interface{}) bool {
	return left.(float64) < right.(float64)
}

func (t float64TypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	return left.(float64) >= right.(float64)
}

func (t float64TypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	return left.(float64) <= right.(float64)
}

// ================ DURATION ====================

type durationTypeHandler struct{}

func (durationTypeHandler) convert(value string) (result interface{}, err error) {
	return time.ParseDuration(value)
}

func (d durationTypeHandler) equal(left interface{}, right interface{}) bool {
	return left.(time.Duration).Nanoseconds() == right.(time.Duration).Nanoseconds()
}

func (d durationTypeHandler) greater(left interface{}, right interface{}) bool {
	return left.(time.Duration).Nanoseconds() > right.(time.Duration).Nanoseconds()
}

func (d durationTypeHandler) less(left interface{}, right interface{}) bool {
	return left.(time.Duration).Nanoseconds() < right.(time.Duration).Nanoseconds()
}

func (d durationTypeHandler) greaterOrEqual(left interface{}, right interface{}) bool {
	return left.(time.Duration).Nanoseconds() >= right.(time.Duration).Nanoseconds()
}

func (d durationTypeHandler) lessOrEqual(left interface{}, right interface{}) bool {
	return left.(time.Duration).Nanoseconds() <= right.(time.Duration).Nanoseconds()
}
