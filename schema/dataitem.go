package schema

import (
	"sync"
	"sync/atomic"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
)

const (
	// FlagData 0 (默认) 数据不压缩, 1 压缩
	FlagData FlagType = iota
	FlagZstd
)

var defaultPool = &Pool{
	pool: sync.Pool{
		New: func() interface{} {
			return new(DataItem)
		},
	},
}

type FlagType int32

type Pool struct {
	pool sync.Pool
}

// Make 获取空数据项
func (p *Pool) Make() *DataItem {
	v := p.pool.Get().(*DataItem)
	v.Body = nil
	v.MarkReset()
	v.FlagReset()
	return v
}

// Release 释放数据项对象, 只回收一次
func (p *Pool) Release(d *DataItem) bool {
	if d.MarkSwap() == 0 {
		d.Reset()
		bytespool.Release(d.Body)
		p.pool.Put(d)
		return true
	}
	return false
}

// Get 获取空数据项
func (p *Pool) Get() *DataItem {
	return p.Make()
}

// Put 释放数据项对象, 只回收一次
func (p *Pool) Put(d *DataItem) {
	p.Release(d)
}

// New 新数据项, 当数据源本身不会变化且 body 不会被用到其他场景时, 以减少数据拷贝
func New(apiname, ip string, body []byte) *DataItem {
	d := defaultPool.Make()
	d.APIName = apiname
	d.IP = ip
	d.Body = body
	return d
}

// NewSafe 新数据项, 复制源数据
func NewSafe(apiname, ip string, body []byte) *DataItem {
	d := defaultPool.Make()
	d.APIName = utils.CopyString(apiname)
	d.IP = utils.CopyString(ip)
	d.Body = bytespool.NewBytes(body)
	return d
}

// NewSafeBody 新数据项, 复制源数据的 body 项
func NewSafeBody(apiname, ip string, body []byte) *DataItem {
	d := defaultPool.Make()
	d.APIName = apiname
	d.IP = ip
	d.Body = bytespool.NewBytes(body)
	return d
}

func Make() *DataItem {
	return defaultPool.Make()
}

func Release(d *DataItem) bool {
	return defaultPool.Release(d)
}

// MarkInc Mark 数值加 1
func (d *DataItem) MarkInc() {
	d.MarkAdd(1)
}

// MarkDec Mark 数值减 1
func (d *DataItem) MarkDec() {
	d.MarkAdd(-1)
}

// MarkAdd Mark 数值增加
func (d *DataItem) MarkAdd(delta int32) {
	atomic.AddInt32(&d.Mark, delta)
}

// MarkValue Mark 返回值
func (d *DataItem) MarkValue() int32 {
	return atomic.LoadInt32(&d.Mark)
}

// MarkSwap Mark 值 -1
func (d *DataItem) MarkSwap() int32 {
	return atomic.SwapInt32(&d.Mark, d.Mark-1)
}

// MarkReset 清除 Mark
func (d *DataItem) MarkReset() {
	atomic.StoreInt32(&d.Mark, 0)
}

// FlagInc Flag 数值加 1
func (d *DataItem) FlagInc() {
	d.FlagAdd(1)
}

// FlagDec Flag 数值减 1
func (d *DataItem) FlagDec() {
	d.FlagAdd(-1)
}

// FlagAdd Flag 数值增加
func (d *DataItem) FlagAdd(delta int32) {
	atomic.AddInt32(&d.Flag, delta)
}

// FlagValue Flag 返回值
func (d *DataItem) FlagValue() int32 {
	return atomic.LoadInt32(&d.Flag)
}

// FlagSwap Flag 值 -1
func (d *DataItem) FlagSwap() int32 {
	return atomic.SwapInt32(&d.Flag, d.Flag-1)
}

// FlagReset 清除 Flag
func (d *DataItem) FlagReset() {
	atomic.StoreInt32(&d.Flag, 0)
}

// Reset 清空数据项
func (d *DataItem) Reset() {
	d.APIName = ""
	d.IP = ""
	d.Body = d.Body[:0]
}

// Release 释放自身
func (d *DataItem) Release() bool {
	return defaultPool.Release(d)
}

func (d *DataItem) String() string {
	return d.APIName + ", " + d.IP + ", " + utils.B2S(d.Body)
}
