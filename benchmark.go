package xtest

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"
)

type B struct {
	memStats        runtime.MemStats
	N               int
	benchFunc       func(b *B)
	bytes           int64
	missingBytes    bool // one of the subbenchmarks does not have bytes set.
	timerOn         bool
	showAllocResult bool
	result          BenchmarkResult
	startAllocs     uint64
	startBytes      uint64
	netAllocs       uint64
	netBytes        uint64
	extra           map[string]float64
	start           time.Time
	duration        time.Duration
}

// StartTimer starts timing a test. This function is called automatically
// before a benchmark starts, but it can also be used to resume timing after
// a call to StopTimer.
func (b *B) StartTimer() {
	if !b.timerOn {
		runtime.ReadMemStats(&b.memStats)
		b.startAllocs = b.memStats.Mallocs
		b.startBytes = b.memStats.TotalAlloc
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
		runtime.ReadMemStats(&b.memStats)
		b.netAllocs += b.memStats.Mallocs - b.startAllocs
		b.netBytes += b.memStats.TotalAlloc - b.startBytes
		b.timerOn = false
	}
}

// ResetTimer zeroes the elapsed benchmark time and memory allocation counters
// and deletes user-reported metrics.
// It does not affect whether the timer is running.
func (b *B) ResetTimer() {
	if b.extra == nil {
		// Allocate the extra map before reading memory stats.
		// Pre-size it to make more allocation unlikely.
		b.extra = make(map[string]float64, 16)
	} else {
		for k := range b.extra {
			delete(b.extra, k)
		}
	}
	if b.timerOn {
		runtime.ReadMemStats(&b.memStats)
		b.startAllocs = b.memStats.Mallocs
		b.startBytes = b.memStats.TotalAlloc
		b.start = time.Now()
	}
	b.duration = 0
	b.netAllocs = 0
	b.netBytes = 0
}

// SetBytes records the number of bytes processed in a single operation.
// If this is called, the benchmark will report ns/op and MB/s.
func (b *B) SetBytes(n int64) { b.bytes = n }

// ReportAllocs enables malloc statistics for this benchmark.
// It is equivalent to setting -test.benchmem, but it only affects the
// benchmark function that calls ReportAllocs.
func (b *B) ReportAllocs() {
	b.showAllocResult = true
}

// launch launches the benchmark function. It gradually increases the number
// of benchmark iterations until the benchmark runs for the requested benchtime.
// launch is run by the doBench function as a separate goroutine.
// run1 must have been called on b.
func (b *B) launch() {
	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.benchFunc(b)
	}
	b.result = BenchmarkResult{b.N, b.duration, b.bytes, b.netAllocs, b.netBytes, b.extra}
}

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	N         int           // The number of iterations.
	T         time.Duration // The total time taken.
	Bytes     int64         // Bytes processed in one iteration.
	MemAllocs uint64        // The total number of memory allocations.
	MemBytes  uint64        // The total number of bytes allocated.

	// Extra records additional metrics reported by ReportMetric.
	Extra map[string]float64
}

// NsPerOp returns the "ns/op" metric.
func (r BenchmarkResult) NsPerOp() int64 {
	if v, ok := r.Extra["ns/op"]; ok {
		return int64(v)
	}
	if r.N <= 0 {
		return 0
	}
	return r.T.Nanoseconds() / int64(r.N)
}

// mbPerSec returns the "MB/s" metric.
func (r BenchmarkResult) mbPerSec() float64 {
	if v, ok := r.Extra["MB/s"]; ok {
		return v
	}
	if r.Bytes <= 0 || r.T <= 0 || r.N <= 0 {
		return 0
	}
	return (float64(r.Bytes) * float64(r.N) / 1e6) / r.T.Seconds()
}

// AllocsPerOp returns the "allocs/op" metric,
// which is calculated as r.MemAllocs / r.N.
func (r BenchmarkResult) AllocsPerOp() int64 {
	if v, ok := r.Extra["allocs/op"]; ok {
		return int64(v)
	}
	if r.N <= 0 {
		return 0
	}
	return int64(r.MemAllocs) / int64(r.N)
}

// AllocedBytesPerOp returns the "B/op" metric,
// which is calculated as r.MemBytes / r.N.
func (r BenchmarkResult) AllocedBytesPerOp() int64 {
	if v, ok := r.Extra["B/op"]; ok {
		return int64(v)
	}
	if r.N <= 0 {
		return 0
	}
	return int64(r.MemBytes) / int64(r.N)
}

// String returns a summary of the benchmark results.
// It follows the benchmark result line format from
// https://golang.org/design/14313-benchmark-format, not including the
// benchmark name.
// Extra metrics override built-in metrics of the same name.
// String does not include allocs/op or B/op, since those are reported
// by MemString.
func (r BenchmarkResult) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%8d", r.N)

	// Get ns/op as a float.
	ns, ok := r.Extra["ns/op"]
	if !ok {
		ns = float64(r.T.Nanoseconds()) / float64(r.N)
	}
	if ns != 0 {
		buf.WriteByte('\t')
		prettyPrint(buf, ns, "ns/op")
	}

	if mbs := r.mbPerSec(); mbs != 0 {
		fmt.Fprintf(buf, "\t%7.2f MB/s", mbs)
	}

	// Print extra metrics that aren't represented in the standard
	// metrics.
	var extraKeys []string
	for k := range r.Extra {
		switch k {
		case "ns/op", "MB/s", "B/op", "allocs/op":
			// Built-in metrics reported elsewhere.
			continue
		}
		extraKeys = append(extraKeys, k)
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		buf.WriteByte('\t')
		prettyPrint(buf, r.Extra[k], k)
	}
	return buf.String()
}

func prettyPrint(w io.Writer, x float64, unit string) {
	// Print all numbers with 10 places before the decimal point
	// and small numbers with three sig figs.
	var format string
	switch y := math.Abs(x); {
	case y == 0 || y >= 99.95:
		format = "%10.0f %s"
	case y >= 9.995:
		format = "%12.1f %s"
	case y >= 0.9995:
		format = "%13.2f %s"
	case y >= 0.09995:
		format = "%14.3f %s"
	case y >= 0.009995:
		format = "%15.4f %s"
	case y >= 0.0009995:
		format = "%16.5f %s"
	default:
		format = "%17.6f %s"
	}
	fmt.Fprintf(w, format, x, unit)
}

// MemString returns r.AllocedBytesPerOp and r.AllocsPerOp in the same format as 'go test'.
func (r BenchmarkResult) MemString() string {
	return fmt.Sprintf("%8d B/op\t%8d allocs/op",
		r.AllocedBytesPerOp(), r.AllocsPerOp())
}

// Benchmark benchmarks a single function. It is useful for creating
// custom benchmarks that do not use the "go test" command.
//
// If f depends on testing flags, then Init must be used to register
// those flags before calling Benchmark and before calling flag.Parse.
//
// If f calls Run, the result will be an estimate of running all its
// subbenchmarks that don't call Run in sequence in a single benchmark.
func Benchmark(n int, f func(b *B)) BenchmarkResult {
	b := &B{
		benchFunc: f,
		N:         n,
		timerOn:   true,
	}
	b.launch()
	return b.result
}
