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

	b := NewSafe(testAPIName, testIP, testBody)
	utils.AssertEqual(t, true, fmt.Sprintf("%p", a) != fmt.Sprintf("%p", b), "&a!=&b")

	// Disable GC to test re-acquire the same data
	gc := debug.SetGCPercent(-1)

	utils.AssertEqual(t, true, a.Release(), "want true")

	c := NewSafe(testAPIName, testIP, testBody)
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
	utils.AssertEqual(t, 512, cap(c.Body))

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

	// 非 Immutable 模式
	g := New(testAPIName, testIP, testBody)
	utils.AssertEqual(t, 510, cap(g.Body))
}

func TestNewSafe(t *testing.T) {
	bs := []byte("unsafe")
	s := utils.B2S(bs)
	item := New(s, s, bs)
	bs[0] = 'x'
	utils.AssertEqual(t, false, item.APIName == "unsafe")
	utils.AssertEqual(t, false, item.IP == "unsafe")
	utils.AssertEqual(t, false, string(item.Body) == "unsafe")

	bs = []byte("safe")
	s = string(bs)
	item = New(s, s, utils.CopyBytes(bs))
	bs[0] = 'x'
	utils.AssertEqual(t, true, item.APIName == "safe")
	utils.AssertEqual(t, true, item.IP == "safe")
	utils.AssertEqual(t, true, string(item.Body) == "safe")

	bs = []byte("safe")
	s = string(bs)
	item = NewSafeBody(s, s, bs)
	bs[0] = 'x'
	utils.AssertEqual(t, true, item.APIName == "safe")
	utils.AssertEqual(t, true, item.IP == "safe")
	utils.AssertEqual(t, true, string(item.Body) == "safe")

	bs = []byte("safe")
	s = utils.B2S(bs)
	item = NewSafe(s, s, bs)
	bs[0] = 'x'
	utils.AssertEqual(t, true, item.APIName == "safe")
	utils.AssertEqual(t, true, item.IP == "safe")
	utils.AssertEqual(t, true, string(item.Body) == "safe")
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
	b.Run("pool.NewSafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := NewSafe(testAPIName, testIP, testBody)
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
	b.Run("pool.NewSafe", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				x := NewSafe(testAPIName, testIP, testBody)
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
// BenchmarkDataItemMake/pool.Make-4                       34401480                35.12 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.Make-4                       34370470                35.17 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.New-4                        27847016                43.22 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.New-4                        28216743                42.91 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool.NewSafe-4                    11295474                106.0 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItemMake/pool.NewSafe-4                    11357240                106.4 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItemMakeParallel/pool.Make-4              133666057                9.106 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.Make-4              133384324                9.018 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.New-4               100000000                11.03 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.New-4               100000000                11.05 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool.NewSafe-4            40727296                28.57 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItemMakeParallel/pool.NewSafe-4            41443420                28.44 ns/op           16 B/op          2 allocs/op
