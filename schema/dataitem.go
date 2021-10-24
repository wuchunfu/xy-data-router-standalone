package schema

import (
	"sync"
	"sync/atomic"

	"github.com/fufuok/utils"
)

type Pool struct {
	capLimit uint64
	pool     sync.Pool
}

var defaultPool = &Pool{
	capLimit: 8192,
	pool: sync.Pool{
		New: func() interface{} {
			return new(DataItem)
		},
	},
}

// Get 获取空数据项
func (p *Pool) Get() *DataItem {
	v := p.pool.Get().(*DataItem)
	v.MarkReset()
	v.FlagReset()
	return v
}

// Put 释放数据项对象, 只回收一次
func (p *Pool) Put(d *DataItem) {
	if d.MarkSwap() == 0 {
		capLimit := int(atomic.LoadUint64(&p.capLimit))
		if capLimit == 0 || cap(d.Body) <= capLimit {
			d.Reset()
			p.pool.Put(d)
		}
	}
}

// SetCapLimit 设置 Body 容量最大可被回收值
func (p *Pool) SetCapLimit(n uint64) {
	atomic.StoreUint64(&p.capLimit, n)
}

// New 新数据项, Immutable
func New(apiname, ip string, body []byte) *DataItem {
	d := Make()
	d.APIName = utils.CopyString(apiname)
	d.Body = utils.CopyBytes(body)
	d.IP = utils.CopyString(ip)
	return d
}

func Make() *DataItem {
	return defaultPool.Get()
}

func Release(d *DataItem) {
	defaultPool.Put(d)
}

func SetCapLimit(n uint64) {
	defaultPool.SetCapLimit(n)
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
func (d *DataItem) Release() {
	defaultPool.Put(d)
}
