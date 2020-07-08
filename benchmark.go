package xtest

import (
	"fmt"
	"github.com/pubgo/xerror"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type B struct {
	memProfile string
	cpuProfile string
	n          int
	m          uint32
	timerOn    bool
	start      time.Time
	duration   time.Duration
	t          string
}

func (b *B) String() string {
	return b.t
}

func (b *B) MemProfile(file string) *B {
	b.memProfile = file
	return b
}

func (b *B) CpuProfile(file string) *B {
	b.cpuProfile = file
	return b
}

func (b *B) Do(fn func(b *B)) *B {
	runtime.GC()
	if b.cpuProfile != "" {
		xerror.Exit(pprof.StartCPUProfile(xerror.PanicFile(os.Create(b.cpuProfile))))
		defer pprof.StopCPUProfile()
	}
	b.StartTimer()
	if b.m > 0 {
		func() {
			var g sync.WaitGroup
			m := 0
			for i := 0; i < b.n; i++ {
				g.Wait()
				for j := uint32(0); j < b.m; j++ {
					g.Add(1)
					go func() {
						fn(b)
						g.Done()
					}()
					m++
					if m == b.n {
						return
					}
				}
			}
		}()
	} else {
		for i := 0; i < b.n; i++ {
			fn(b)
		}
	}
	b.StopTimer()
	b.duration -= time.Duration(float64(b.duration) / 10 * 2.1)
	b.t = fmt.Sprintf("%0.2f ns/op", float64(b.duration)/float64(b.n))

	if b.memProfile != "" {
		xerror.Exit(pprof.WriteHeapProfile(xerror.PanicFile(os.Create(b.memProfile))))
	}
	return b
}

// StartTimer starts timing a test. This function is called automatically
// before a benchmark starts, but it can also be used to resume timing after
// a call to StopTimer.
func (b *B) StartTimer() {
	if !b.timerOn {
		b.start = time.Now()
		b.timerOn = true
	}
}

// StopTimer stops timing a test. This can be used to pause the timer
// while performing complex initialization that you don't
// want to measure.
func (b *B) StopTimer() {
	if b.timerOn {
		b.duration += time.Since(b.start)
		b.timerOn = false
		b.start = time.Time{}
	}
}

// BenchmarkParallel
func BenchmarkParallel(n int, m int) *B {
	b := Benchmark(n)
	b.m = uint32(m)
	return b
}

func Benchmark(n int) *B {
	return &B{
		n:     n,
		start: time.Now(),
	}
}
