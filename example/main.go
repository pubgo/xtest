package main

import (
	"fmt"
	"github.com/pubgo/xtest"
)

func main() {
	xtest.Debug("", fmt.Println, xtest.Debug)
	xtest.Debugln("", fmt.Println, xtest.Debug)
	xtest.Benchmark(100).
		Do(func(b xtest.B) {
			b.StopTimer()
			b.StartTimer()
		})
}
