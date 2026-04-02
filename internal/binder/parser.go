package binder

import (
	"math"
	"strconv"

	"github.com/google/uuid"
)

// Exported parsing helpers so other packages (eg. routing) can reuse the
// same conversions and avoid duplication.

func ParseIntVal(val string) (int, bool) {
	n64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}
	if n64 < int64(math.MinInt) || n64 > int64(math.MaxInt) {
		return 0, false
	}
	return int(n64), true
}

func ParseIntSlice(vals []string) ([]int, bool) {
	return parseSlice(vals, func(s string) (int, error) {
		n64, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		if n64 < int64(math.MinInt) || n64 > int64(math.MaxInt) {
			return 0, strconv.ErrRange
		}
		return int(n64), nil
	})
}

func ParseInt16Val(val string) (int16, bool) {
	n64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}
	if n64 < math.MinInt16 || n64 > math.MaxInt16 {
		return 0, false
	}
	return int16(n64), true
}

func ParseInt16Slice(vals []string) ([]int16, bool) {
	return parseSlice(vals, func(s string) (int16, error) {
		n64, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		if n64 < math.MinInt16 || n64 > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(n64), nil
	})
}

func ParseInt32Val(val string) (int32, bool) {
	n64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}
	if n64 < math.MinInt32 || n64 > math.MaxInt32 {
		return 0, false
	}
	return int32(n64), true
}

func ParseInt32Slice(vals []string) ([]int32, bool) {
	return parseSlice(vals, func(s string) (int32, error) {
		n64, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		if n64 < math.MinInt32 || n64 > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int32(n64), nil
	})
}

func ParseInt64Val(val string) (int64, bool) {
	n, err := strconv.ParseInt(val, 10, 64)
	return n, err == nil
}

func ParseInt64Slice(vals []string) ([]int64, bool) {
	return parseSlice(vals, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
}

func ParseBoolVal(val string) (bool, bool) {
	b, err := strconv.ParseBool(val)
	return b, err == nil
}

func ParseBoolSlice(vals []string) ([]bool, bool) {
	return parseSlice(vals, strconv.ParseBool)
}

func ParseFloat32Val(val string) (float32, bool) {
	f, err := strconv.ParseFloat(val, 32)
	return float32(f), err == nil
}

func ParseFloat32Slice(vals []string) ([]float32, bool) {
	return parseSlice(vals, func(s string) (float32, error) {
		f, err := strconv.ParseFloat(s, 32)
		return float32(f), err
	})
}

func ParseFloat64Val(val string) (float64, bool) {
	f, err := strconv.ParseFloat(val, 64)
	return f, err == nil
}

func ParseFloat64Slice(vals []string) ([]float64, bool) {
	return parseSlice(vals, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
}

// UUID helpers
func ParseUUIDVal(val string) (uuid.UUID, bool) {
	u, err := uuid.Parse(val)
	return u, err == nil
}

func ParseUUIDSlice(vals []string) ([]uuid.UUID, bool) {
	return parseSlice(vals, uuid.Parse)
}
