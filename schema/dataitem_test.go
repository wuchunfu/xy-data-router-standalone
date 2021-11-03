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

	a.Release()

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
	b.Release()

	e := Make()
	utils.AssertEqual(t, true, fmt.Sprintf("%p", e) != fmt.Sprintf("%p", b), "&e!=&b")

	b.Release()

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
// BenchmarkDataItemMake/pool-4              33834816                35.54 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool-4              33093555                35.25 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/new-4               27697900                41.84 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMake/new-4               27687626                42.25 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMakeParallel/pool-4      27874159                43.08 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool-4      27840216                43.02 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/new-4       28916372                39.55 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMakeParallel/new-4       30574990                39.75 ns/op           64 B/op          1 allocs/op
