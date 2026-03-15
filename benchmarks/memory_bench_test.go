package benchmarks

import (
	"runtime"
	"testing"
)

func BenchmarkMemoryUsage(b *testing.B) {
	store := setupIndexedStore(b, 100)
	defer store.Close()

	// Force GC to get clean measurement
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	b.ReportMetric(float64(m.Alloc)/1024/1024, "MB-alloc")
	b.ReportMetric(float64(m.Sys)/1024/1024, "MB-sys")
	b.ReportMetric(float64(m.HeapInuse)/1024/1024, "MB-heap")
}
