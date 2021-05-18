package gokql

import (
	"strconv"
	"time"
)

type typeHandler interface {
	convert(value string) (result interface{}, err error)
	equal(left interface{}, right interface{}) (result bool, ok bool)
	greater(left interface{}, right interface{}) (result bool, ok bool)
	less(left interface{}, right interface{}) (result bool, ok bool)
	greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool)
	lessOrEqual(left interface{}, right interface{}) (result bool, ok bool)
}

// ================ INT64  ====================

type int64TypeHandler struct{}

func (int64TypeHandler) compare(left interface{}, right interface{}, operation func(int64, int64) bool) (result bool, ok bool) {
	var leftInt, rightInt int64
	if leftInt, ok = left.(int64); !ok {
		return false, false
	}
	if rightInt, ok = right.(int64); !ok {
		return false, false
	}
	return operation(leftInt, rightInt), true
}

func (int64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseInt(value, 10, 64)
}

func (i int64TypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r int64, l int64) bool { return l == r })
}

func (i int64TypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r int64, l int64) bool { return l > r })
}

func (i int64TypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r int64, l int64) bool { return l < r })
}

func (i int64TypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r int64, l int64) bool { return l >= r })
}

func (i int64TypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r int64, l int64) bool { return l <= r })
}

// ================ INT64  ====================

type uint64TypeHandler struct{}

func (uint64TypeHandler) compare(left interface{}, right interface{}, operation func(uint64, uint64) bool) (result bool, ok bool) {
	var leftInt, rightInt uint64
	if leftInt, ok = left.(uint64); !ok {
		return false, false
	}
	if rightInt, ok = right.(uint64); !ok {
		return false, false
	}
	return operation(leftInt, rightInt), true
}

func (uint64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseUint(value, 10, 64)
}

func (i uint64TypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r uint64, l uint64) bool { return l == r })
}

func (i uint64TypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r uint64, l uint64) bool { return l > r })
}

func (i uint64TypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r uint64, l uint64) bool { return l < r })
}

func (i uint64TypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r uint64, l uint64) bool { return l >= r })
}

func (i uint64TypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return i.compare(left, right, func(r uint64, l uint64) bool { return l <= r })
}

// ================ TIME  ====================

type timeTypeHandler struct{}

func (timeTypeHandler) compare(left interface{}, right interface{}, operation func(time.Time, time.Time) bool) (result bool, ok bool) {
	var leftVal, rightVal time.Time
	if leftVal, ok = left.(time.Time); !ok {
		return false, false
	}
	if rightVal, ok = right.(time.Time); !ok {
		return false, false
	}
	return operation(leftVal, rightVal), true
}

func (timeTypeHandler) convert(value string) (result interface{}, err error) {
	return time.Parse(time.RFC3339, value)
}

func (t timeTypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r time.Time, l time.Time) bool { return l.Equal(r) })
}

func (t timeTypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r time.Time, l time.Time) bool { return l.After(r) })
}

func (t timeTypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r time.Time, l time.Time) bool { return l.Before(r) })
}

func (t timeTypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r time.Time, l time.Time) bool { return l.After(r) || l.Equal(r) })
}

func (t timeTypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r time.Time, l time.Time) bool { return l.Before(r) || l.Equal(r) })
}

// ================ BOOL  ====================

type boolTypeHandler struct{}

func (boolTypeHandler) compare(left interface{}, right interface{}, operation func(bool, bool) bool) (result bool, ok bool) {
	var leftVal, rightVal bool
	if leftVal, ok = left.(bool); !ok {
		return false, false
	}
	if rightVal, ok = right.(bool); !ok {
		return false, false
	}
	return operation(leftVal, rightVal), true
}

func (boolTypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseBool(value)
}

func (t boolTypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r bool, l bool) bool { return l == r })
}

func (t boolTypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r bool, l bool) bool { return l && !r })
}

func (t boolTypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r bool, l bool) bool { return !l && r })
}

func (t boolTypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r bool, l bool) bool { return (l && !r) || (l == r) })
}

func (t boolTypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r bool, l bool) bool { return (!l && r) || (l == r) })
}

// ================ FLOAT ====================

type float64TypeHandler struct{}

func (float64TypeHandler) compare(left interface{}, right interface{}, operation func(float64, float64) bool) (result bool, ok bool) {
	var leftVal, rightVal float64
	if leftVal, ok = left.(float64); !ok {
		return false, false
	}
	if rightVal, ok = right.(float64); !ok {
		return false, false
	}
	return operation(leftVal, rightVal), true
}

func (float64TypeHandler) convert(value string) (result interface{}, err error) {
	return strconv.ParseBool(value)
}

func (t float64TypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r float64, l float64) bool { return l == r })
}

func (t float64TypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r float64, l float64) bool { return l > r })
}

func (t float64TypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r float64, l float64) bool { return l < r })
}

func (t float64TypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r float64, l float64) bool { return l >= r })
}

func (t float64TypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return t.compare(left, right, func(r float64, l float64) bool { return l <= r })
}

// ================ DURATION ====================

type durationTypeHandler struct{}

func (durationTypeHandler) compare(left interface{}, right interface{}, operation func(time.Duration, time.Duration) bool) (result bool, ok bool) {
	var leftVal, rightVal time.Duration
	if leftVal, ok = left.(time.Duration); !ok {
		return false, false
	}
	if rightVal, ok = right.(time.Duration); !ok {
		return false, false
	}
	return operation(leftVal, rightVal), true
}

func (durationTypeHandler) convert(value string) (result interface{}, err error) {
	return time.ParseDuration(value)
}

func (d durationTypeHandler) equal(left interface{}, right interface{}) (result bool, ok bool) {
	return d.compare(left, right, func(r time.Duration, l time.Duration) bool { return l.Nanoseconds() == r.Nanoseconds() })
}

func (d durationTypeHandler) greater(left interface{}, right interface{}) (result bool, ok bool) {
	return d.compare(left, right, func(r time.Duration, l time.Duration) bool { return l.Nanoseconds() > r.Nanoseconds() })
}

func (d durationTypeHandler) less(left interface{}, right interface{}) (result bool, ok bool) {
	return d.compare(left, right, func(r time.Duration, l time.Duration) bool { return l.Nanoseconds() < r.Nanoseconds() })
}

func (d durationTypeHandler) greaterOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return d.compare(left, right, func(r time.Duration, l time.Duration) bool { return l.Nanoseconds() >= r.Nanoseconds() })
}

func (d durationTypeHandler) lessOrEqual(left interface{}, right interface{}) (result bool, ok bool) {
	return d.compare(left, right, func(r time.Duration, l time.Duration) bool { return l.Nanoseconds() <= r.Nanoseconds() })
}
