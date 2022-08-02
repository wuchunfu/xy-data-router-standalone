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
	tunChan *chanx.UnboundedChan[*schema.DataItem]

	// ItemTotal Tunnel 服务端接收和客户端发送计数
	ItemTotal     xsync.Counter
	RecvCount     xsync.Counter
	RecvBadCount  xsync.Counter
	SendCount     xsync.Counter
	SendErrors    xsync.Counter
	CompressTotal xsync.Counter
)

// InitMain 程序启动时初始化
func InitMain() {
	// 初始化 Tun 数据信道
	tunChan = common.NewChanx[*schema.DataItem]()

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
		item := item
		ItemTotal.Inc()
		// 设置压缩标识
		if item.Size() >= conf.Config.TunConf.CompressMinSize {
			item.Flag = 1
			CompressTotal.Inc()
		}
		tunChan.In <- item
	}
	common.Log.Error().Msg("Exception: Tunnel entry worker exited")
}
