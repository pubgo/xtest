package xtest

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RangeBytes ...
func MockBytes(min, max int) []byte {
	var dt = make([]byte, MockInt(min, max))
	rand.Read(dt)
	return dt
}

// RangeString ...
func MockString(min, max int) string {
	return string(MockBytes(min, max))
}

// RangeDur ...
func MockDur(min, max time.Duration) time.Duration {
	return time.Duration(MockInt(int(min), int(max)))
}

// Range ...
func MockInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + rand.Intn(max-min)
}
