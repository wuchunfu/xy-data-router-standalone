package schema

import (
	"github.com/fufuok/chanx"
	"github.com/fufuok/utils/xsync"

	"github.com/fufuok/xy-data-router/common"
)

var (
	ItemDrChan  *chanx.UnboundedChan[*DataItem]
	ItemTunChan *chanx.UnboundedChan[*DataItem]
	ItemTotal   xsync.Counter
)

// InitMain 程序启动时初始化配置
func InitMain() {
	ItemDrChan = common.NewChanx[*DataItem]()
	ItemTunChan = common.NewChanx[*DataItem]()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {}

func Stop() {}

// PushDataToChanx 接收数据推入队列
func PushDataToChanx(item *DataItem) {
	ItemTotal.Inc()
	if common.ForwardTunnel != "" {
		ItemTunChan.In <- item
		return
	}
	ItemDrChan.In <- item
}
