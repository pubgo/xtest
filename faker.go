package xtest

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RangeBytes ...
func RangeBytes(min, max int) []byte {
	var dt = make([]byte, Range(min, max))
	rand.Read(dt)
	return dt
}

// RangeString ...
func RangeString(min, max int) string {
	return string(RangeBytes(min, max))
}

// RangeDur ...
func RangeDur(min, max time.Duration) time.Duration {
	return time.Duration(Range(int(min), int(max)))
}

// Range ...
func Range(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}
