package xtest

import (
	"fmt"
	"testing"
	"time"
)

func TestBen(t *testing.T) {
	MemStatsPrint()
	bm := Benchmark(1000).
		Do(func(b B) {
			time.Sleep(time.Millisecond)
		})
	MemStatsPrint()
	fmt.Println(bm)
	Debugln()
}
