package xtest

import (
	"fmt"
	"github.com/pubgo/xerror"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
	"unsafe"
)

var defaultMemStats = initStats()

const memStatsSize = uint64(unsafe.Sizeof(runtime.MemStats{}))

func initStats() (m runtime.MemStats) {
	runtime.ReadMemStats(&m)
	return
}

type B struct {
	startStats runtime.MemStats
	stopStats  runtime.MemStats
	memProfile string
	cpuProfile string
	N          int
	M          uint32
	timerOn    bool
	constBytes int
	allocBytes int
	start      time.Time
	duration   time.Duration
	T          string
}

func (b *B) String() string {
	return b.T
}

func (b *B) AllocBytes() uint64{
	return uint64(b.allocBytes)
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
	if b.M > 0 {
		func() {
			var g sync.WaitGroup
			m := 0
			for i := 0; i < b.N; i++ {
				g.Wait()
				for j := uint32(0); j < b.M; j++ {
					g.Add(1)
					go func() {
						fn(b)
						g.Done()
					}()
					m++
					if m == b.N {
						return
					}
				}
			}
		}()

	} else {
		for i := 0; i < b.N; i++ {
			fn(b)
		}
	}
	b.StopTimer()
	b.allocBytes -= b.constBytes
	b.duration -= time.Duration(float64(b.duration) / 10 * 2.1)
	b.T = fmt.Sprintf("%0.2f ns/op", float64(b.duration)/float64(b.N))

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
		runtime.ReadMemStats(&b.startStats)
		b.constBytes = int(b.startStats.HeapSys - b.stopStats.HeapSys)
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
		runtime.ReadMemStats(&b.stopStats)
		b.allocBytes += int(b.startStats.HeapSys - b.stopStats.HeapSys - memStatsSize)
		b.timerOn = false
		b.start = time.Time{}
	}
}

// Benchmark benchmarks a single function. It is useful for creating
// custom benchmarks that do not use the "go test" command.
//
// If f depends on testing flags, then Init must be used to register
// those flags before calling Benchmark and before calling flag.Parse.
//
// If f calls Run, the result will be an estimate of running all its
// subbenchmarks that don't call Run in sequence in a single benchmark.
func BenchmarkParallel(n int, m int) *B {
	b := Benchmark(n)
	b.M = uint32(m)
	return b
}

func Benchmark(n int) *B {
	return &B{
		N:         n,
		stopStats: defaultMemStats,
		start:     time.Now(),
	}
}
