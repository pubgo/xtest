package xtest

import (
	"fmt"
	"runtime"
)

func MemStatsPrint() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc = %v HeapIdel= %v HeapSys = %v  HeapReleased = %v\n", m.HeapAlloc/1024, m.HeapIdle/1024, m.HeapSys/1024, m.HeapReleased/1024)
}
