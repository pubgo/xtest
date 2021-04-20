package xtest

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/pubgo/xerror"
)

type B interface {
	StartTimer()
	StopTimer()
}

type IBenchmark interface {
	CpuProfile(file string) IBenchmark
	MemProfile(file string) IBenchmark
	Do(fn func(b B)) IBenchmark
	String() string
}

type benchmark struct {
	memProfile string
	cpuProfile string
	n          int
	m          uint32
	timerOn    bool
	start      time.Time
	duration   time.Duration
	t          string
}

func (b *benchmark) String() string {
	return b.t
}

func (b *benchmark) MemProfile(file string) IBenchmark {
	b.memProfile = file
	return b
}

func (b *benchmark) CpuProfile(file string) IBenchmark {
	b.cpuProfile = file
	return b
}

func (b *benchmark) Do(fn func(b B)) IBenchmark {
	runtime.GC()
	if b.cpuProfile != "" {
		xerror.Exit(pprof.StartCPUProfile(xerror.PanicErr(os.Create(b.cpuProfile)).(*os.File)))
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
		xerror.Exit(pprof.WriteHeapProfile(xerror.PanicErr(os.Create(b.memProfile)).(*os.File)))
	}
	return b
}

// StartTimer starts timing a test. This function is called automatically
// before a benchmark starts, but it can also be used to resume timing after
// a call to StopTimer.
func (b *benchmark) StartTimer() {
	if !b.timerOn {
		b.start = time.Now()
		b.timerOn = true
	}
}

// StopTimer stops timing a test. This can be used to pause the timer
// while performing complex initialization that you don't
// want to measure.
func (b *benchmark) StopTimer() {
	if b.timerOn {
		b.duration += time.Since(b.start)
		b.timerOn = false
		b.start = time.Time{}
	}
}

// BenchmarkParallel
func BenchmarkParallel(n int, m int) IBenchmark {
	return &benchmark{
		m:     uint32(m),
		n:     n,
		start: time.Now(),
	}
}

func Benchmark(n int) IBenchmark {
	return &benchmark{
		n:     n,
		start: time.Now(),
	}
}
