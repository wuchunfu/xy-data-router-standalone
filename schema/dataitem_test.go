package schema

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/fufuok/utils"
)

var (
	testAPIName = "test"
	testIP      = "192.168.1.1"
	testBody    = utils.FastRandBytes(510)
)

func TestDataItem_Pool(t *testing.T) {
	a := Make()
	utils.AssertEqual(t, true, a.Body == nil)

	b := New(testAPIName, testIP, testBody)
	utils.AssertEqual(t, true, fmt.Sprintf("%p", a) != fmt.Sprintf("%p", b), "&a!=&b")

	// Disable GC to test re-acquire the same data
	gc := debug.SetGCPercent(-1)

	utils.AssertEqual(t, true, a.Release(), "want true")

	c := New(testAPIName, testIP, testBody)
	utils.AssertEqual(t, true, fmt.Sprintf("%p", c) == fmt.Sprintf("%p", a), "&c==&a")

	d := Make()
	utils.AssertEqual(t, true, fmt.Sprintf("%p", d) != fmt.Sprintf("%p", a), "&d!=&a")
	utils.AssertEqual(t, true, fmt.Sprintf("%p", d) != fmt.Sprintf("%p", b), "&d!=&b")
	utils.AssertEqual(t, true, fmt.Sprintf("%p", d) != fmt.Sprintf("%p", c), "&d!=&c")

	b.Reset()
	utils.AssertEqual(t, "", b.IP)
	utils.AssertEqual(t, "", b.APIName)
	utils.AssertEqual(t, 0, len(b.Body))
	utils.AssertEqual(t, 512, cap(b.Body))

	b.MarkInc()
	b.MarkInc()
	b.MarkDec()
	utils.AssertEqual(t, false, b.Release(), "want false")

	e := Make()
	utils.AssertEqual(t, true, fmt.Sprintf("%p", e) != fmt.Sprintf("%p", b), "&e!=&b")

	utils.AssertEqual(t, true, b.Release(), "want true")
	utils.AssertEqual(t, false, b.Release(), "want false")
	utils.AssertEqual(t, false, b.Release(), "want false")

	f := Make()
	utils.AssertEqual(t, true, fmt.Sprintf("%p", f) == fmt.Sprintf("%p", b), "&f==&b")

	// Re-enable GC
	debug.SetGCPercent(gc)

	// 注意, Release 后的变量不要再使用, 不可预料
	c.IP = "7.7.7.7"
	utils.AssertEqual(t, "7.7.7.7", a.IP)
}

func BenchmarkDataItemMake(b *testing.B) {
	b.Run("pool.Make", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := Make()
			x.Release()
		}
	})
	b.Run("pool.New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := New(testAPIName, testIP, testBody)
			x.Release()
		}
	})
}

func BenchmarkDataItemMakeParallel(b *testing.B) {
	b.Run("pool.Make", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				x := Make()
				x.Release()
			}
		})
	})
	b.Run("pool.New", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				x := New(testAPIName, testIP, testBody)
				x.Release()
			}
		})
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkDataItemMake
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/schema
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkDataItemMake/pool.Make-4       34318790                34.91 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.Make-4       34436946                35.01 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.New-4        11614143                101.2 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItemMake/pool.New-4        11971916                101.4 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItemMakeParallel/pool.Make-4               134554711                8.870 ns/op           0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.Make-4               135073802                8.871 ns/op           0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.New-4                 44045688                27.30 ns/op          16 B/op          2 allocs/op
// BenchmarkDataItemMakeParallel/pool.New-4                 44745177                27.89 ns/op          16 B/op          2 allocs/op
