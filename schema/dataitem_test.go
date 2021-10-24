package schema

import (
	"fmt"
	"testing"

	"github.com/fufuok/utils"
)

func TestDataItem_Pool(t *testing.T) {
	apiname := "test"
	ip := "192.168.1.1"
	body := utils.FastRandBytes(512)
	a := Make()
	b := New(apiname, ip, body)
	utils.AssertEqual(t, true, fmt.Sprintf("%p", a) != fmt.Sprintf("%p", b), "&a!=&b")

	a.Release()

	c := New(apiname, ip, body)
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

	SetCapLimit(511)
	f.Release()

	g := Make()
	utils.AssertEqual(t, true, fmt.Sprintf("%p", g) != fmt.Sprintf("%p", f), "&g!=&f")

	// 注意, Release 后的变量不要再使用, 不可预料
	c.IP = "7.7.7.7"
	utils.AssertEqual(t, "7.7.7.7", a.IP)
}

func BenchmarkDataItemMake(b *testing.B) {
	x := Make()
	b.Run("pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x = Make()
			x.Release()
		}
	})
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x = new(DataItem)
		}
	})
}

func BenchmarkDataItemMakeParallel(b *testing.B) {
	x := Make()
	b.Run("pool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				x = Make()
				x.Release()
			}
		})
	})
	b.Run("new", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				x = new(DataItem)
			}
		})
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkDataItemMake
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/schema
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkDataItemMake/pool-4                    33962113                35.48 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/pool-4                    33198151                35.66 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMake/new-4                     27157551                41.87 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMake/new-4                     28881915                42.21 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMakeParallel/pool-4            25939573                46.37 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/pool-4            25888813                46.38 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItemMakeParallel/new-4             27567992                43.88 ns/op           64 B/op          1 allocs/op
// BenchmarkDataItemMakeParallel/new-4             27806358                43.53 ns/op           64 B/op          1 allocs/op
