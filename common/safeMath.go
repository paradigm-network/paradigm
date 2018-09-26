package common

import "math"

const (
	MAX_UINT64 = math.MaxUint64
)

func SafeSub(x, y uint64) (uint64, bool) {
	return x - y, x < y
}

func SafeAdd(x, y uint64) (uint64, bool) {
	return x + y, y > MAX_UINT64-x
}

func SafeMul(x, y uint64) (uint64, bool) {
	if x == 0 || y == 0 {
		return 0, false
	}
	return x * y, y > MAX_UINT64/x
}
