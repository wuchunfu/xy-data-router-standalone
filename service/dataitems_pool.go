package service

import (
	"sync"

	"github.com/fufuok/xy-data-router/schema"
)

var dataItemsPool sync.Pool

// DataItem 数据集
type tDataItems struct {
	items []*schema.DataItem
	size  int
	count int
}

func (dis *tDataItems) add(item *schema.DataItem) {
	dis.items = append(dis.items, item)
	dis.size += len(item.Body)
	dis.count++
}

func (dis *tDataItems) release() {
	putDataItems(dis)
}

func getDataItems() *tDataItems {
	v := dataItemsPool.Get()
	if v != nil {
		return v.(*tDataItems)
	}
	return new(tDataItems)
}

// 回收所有数据项
func putDataItems(dis *tDataItems) {
	for i := range dis.items {
		dis.items[i].Release()
	}
	dis.items = dis.items[:0]
	dis.size = 0
	dis.count = 0
	dataItemsPool.Put(dis)
}
