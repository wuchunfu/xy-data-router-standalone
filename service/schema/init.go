package schema

import (
	"github.com/fufuok/chanx"

	"github.com/fufuok/xy-data-router/common"
)

var (
	ItemDrChan  *chanx.UnboundedChan
	ItemTunChan *chanx.UnboundedChan
)

// InitMain 程序启动时初始化配置
func InitMain() {
	ItemDrChan = common.NewChanx()
	ItemTunChan = common.NewChanx()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {}

func Stop() {}

// PushDataToChanx 接收数据推入队列
func PushDataToChanx(item *DataItem) {
	if common.ForwardTunnel != "" {
		ItemTunChan.In <- item
		return
	}
	ItemDrChan.In <- item
}
