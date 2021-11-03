package service

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
)

// 附加系统字段(旧): 入参 JS 数据必须为 {JSON字典}, Immutable
func testAppendSYSField(js []byte, ip string) []byte {
	if len(js) == 0 {
		return nil
	}
	if gjson.GetBytes(js, "_cip").Exists() {
		return utils.CopyBytes(js)
	}
	return append(
		utils.AddStringBytes(
			`{"_cip":"`, ip,
			`","_ctime":"`, common.Now3399UTC,
			`","_gtime":"`, common.Now3399, `",`,
		),
		js[1:]...,
	)
}

func TestAppendSYSField(t *testing.T) {
	testjs := []byte(`{"ff": "ok"}`)
	testcip := []byte(`{"ff": "ok", "_cip":""}`)
	tests := []struct {
		js []byte
		ip string
	}{
		{nil, ""},
		{nil, "1.1.1.1"},
		{testjs, "255.255.255.255"},
		{testcip, "255.255.255.1"},
		{utils.FastRandBytes(9999), "255.255.255.1"},
	}
	for _, v := range tests {
		t.Run(fmt.Sprintf("ip(%s)", v.ip), func(t *testing.T) {
			want := testAppendSYSField(v.js, v.ip)
			got := appendSYSField(v.js, v.ip)
			utils.AssertEqual(t, true, bytes.Equal(want, got))
			bytespool.Release(got)
		})
	}
}

func BenchmarkAppendSysField_Normal(b *testing.B) {
	ip := "255.255.255.255"
	js := utils.FastRandBytes(256)
	b.ReportAllocs()
	b.ResetTimer()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := appendSYSField(js, ip)
			_ = buf
			bytespool.Release(buf)
		}
	})
	b.Run("old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := testAppendSYSField(js, ip)
			_ = buf
		}
	})
}

func BenchmarkAppendSysField_Larger(b *testing.B) {
	ip := "255.255.255.255"
	// 超出了默认的字节池最大长度 8192
	js := utils.FastRandBytes(9999)
	b.ReportAllocs()
	b.ResetTimer()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := appendSYSField(js, ip)
			_ = buf
			bytespool.Release(buf)
		}
	})
	b.Run("old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := testAppendSYSField(js, ip)
			_ = buf
		}
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkAppendSysField
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/service
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkAppendSysField_Normal/new-4             3755038              330.5 ns/op             17 B/op          0 allocs/op
// BenchmarkAppendSysField_Normal/new-4             3782475              321.9 ns/op             17 B/op          0 allocs/op
// BenchmarkAppendSysField_Normal/old-4             3407594              358.7 ns/op            448 B/op          2 allocs/op
// BenchmarkAppendSysField_Normal/old-4             3414798              353.2 ns/op            448 B/op          2 allocs/op
// BenchmarkAppendSysField_Larger/new-4              183744              6440 ns/op           10240 B/op          1 allocs/op
// BenchmarkAppendSysField_Larger/new-4              186739              6402 ns/op           10240 B/op          1 allocs/op
// BenchmarkAppendSysField_Larger/old-4              187136              6488 ns/op           10336 B/op          2 allocs/op
// BenchmarkAppendSysField_Larger/old-4              189338              6481 ns/op           10336 B/op          2 allocs/op
