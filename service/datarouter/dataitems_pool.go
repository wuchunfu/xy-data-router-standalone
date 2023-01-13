package datarouter

import (
	"sync"

	"github.com/fufuok/xy-data-router/service/schema"
)

var dataItemsPool sync.Pool

// DataItem 数据集
type dataItems struct {
	items []*schema.DataItem
	size  int
	count int
}

func (dis *dataItems) add(item *schema.DataItem) {
	dis.items = append(dis.items, item)
	dis.size += len(item.Body)
	dis.count++
}

func (dis *dataItems) release() {
	putDataItems(dis)
}

func getDataItems() *dataItems {
	v := dataItemsPool.Get()
	if v != nil {
		return v.(*dataItems)
	}
	return new(dataItems)
}

// 回收所有数据项
func putDataItems(dis *dataItems) {
	for i := range dis.items {
		dis.items[i].Release()
	}
	dis.items = dis.items[:0]
	dis.size = 0
	dis.count = 0
	dataItemsPool.Put(dis)
}
