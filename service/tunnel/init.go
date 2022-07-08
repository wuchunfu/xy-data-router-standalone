package tunnel

import (
	"github.com/fufuok/chanx"
	"github.com/fufuok/utils/xsync"
	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

const tunMethod = "tunnel"

var (
	// Tun 数据信道
	tunChan *chanx.UnboundedChan

	// TunRecvCount Tunnel 服务端接收和客户端发送计数
	TunRecvCount     xsync.Counter
	TunRecvBadCount  xsync.Counter
	TunSendCount     xsync.Counter
	TunSendErrors    xsync.Counter
	TunDataTotal     xsync.Counter
	TunCompressTotal xsync.Counter
)

// InitMain 程序启动时初始化
func InitMain() {
	// 初始化 Tun 数据信道
	tunChan = common.NewChanx()

	initLogger()
	arpc.SetSendQueueSize(conf.Config.TunConf.SendQueueSize)
	arpc.EnablePool(true)
	arpc.SetAsyncResponse(true)
	arpc.SetAsyncExecutor(func(f func()) {
		_ = common.GoPool.Submit(f)
	})

	go initTunServer()
	go initTunClient()
	go dataEntry()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	loadLogger()
}

func Stop() {}

// 数据入口
func dataEntry() {
	for item := range schema.ItemTunChan.Out {
		item := item.(*schema.DataItem)
		TunDataTotal.Inc()
		// 设置压缩标识
		if item.Size() >= conf.Config.TunConf.CompressMinSize {
			item.Flag = 1
			TunCompressTotal.Inc()
		}
		tunChan.In <- item
	}
}
