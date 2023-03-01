package datarouter

import (
	"github.com/fufuok/ants"
	"github.com/fufuok/chanx"
	"github.com/fufuok/utils/xsync"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger/alarm"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/service/schema"
)

// JS 数据必须为 JSON 字典, 数据至少包含 1 个键值对, 即最小长为 7, 如: {"L":7}
const jsonMinLen = 7

var (
	// ES 数据分隔符
	esBodySep = []byte(conf.ESBodySep)

	// 以接口名为键的数据通道
	dataRouters = xsync.NewMapOf[*router]()

	// ItemTotal 收到的数据项计数
	ItemTotal xsync.Counter

	// ESChan ES 数据接收信道
	ESChan *chanx.UnboundedChan[*schema.DataItem]

	// DataProcessorDiscards 数据处理丢弃计数, 超过 ProcessorWaitingLimit
	DataProcessorDiscards xsync.Counter

	// ESDataItemDiscards 数据丢弃计数, 繁忙时丢弃可选接口的数据, 不写 ES
	ESDataItemDiscards xsync.Counter

	// ESDataTotal ES 收到数据数量计数
	ESDataTotal xsync.Counter

	// ESBulkDiscards ES Bulk 写入丢弃协程数, 超过 ESBulkerWaitingLimit
	ESBulkDiscards xsync.Counter

	// UDPRequestCount UDP 请求计数
	UDPRequestCount xsync.Counter

	// DataProcessorPool 数据处理协程池
	DataProcessorPool *ants.PoolWithFunc

	// ESBulkPool ES 写入协程池
	ESBulkPool *ants.PoolWithFunc
)

// 数据分发
type router struct {
	// 数据接收信道
	drChan *chanx.UnboundedChan[*schema.DataItem]

	// 接口数据分发信道
	apiChan *chanx.UnboundedChan[*schema.DataItem]

	// 接口配置
	apiConf *conf.APIConf
}

// 数据处理
type processor struct {
	dr   *router
	data *schema.DataItem
}

// InitMain 程序启动时初始化
func InitMain() {
	// 初始化 ES 数据信道
	ESChan = common.NewChanx[*schema.DataItem]()

	initESWriteStatus()
	initDataRouter()

	// 开启 ES 写入
	go esWorker()

	// 启动 UDP 接口服务
	go initUDPServer()

	// ES 索引头信息更新
	go updateESBulkHeader()

	// 初始化数据处理
	go initDataProcessorPool()
	go initESWriteBreaker()
	go initESBulkPool()
	go dataEntry()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	// 配置变化时, 热加载
	initESWriteStatus()
	initDataRouter()

	// 调节协程池
	tuneDataProcessorSize()
	tuneESBulkerSize()
}

func Stop() {
	poolRelease()
}

func poolRelease() {
	DataProcessorPool.Release()
	ESBulkPool.Release()
}

// tuneDataProcessorSize 调节协程并发数, 最大阻塞任务数
func tuneDataProcessorSize() {
	DataProcessorPool.Tune(conf.Config.DataConf.ProcessorSize)
	DataProcessorPool.TuneMaxBlockingTasks(conf.Config.DataConf.ProcessorWaitingLimit)
}

// tuneESBulkerSize 调节协程并发数, 最大阻塞任务数
func tuneESBulkerSize() {
	ESBulkPool.Tune(conf.Config.DataConf.ESBulkerSize)
	ESBulkPool.TuneMaxBlockingTasks(conf.Config.DataConf.ESBulkerWaitingLimit)
}

// 新数据信道
func newDataRouter(apiConf *conf.APIConf) *router {
	return &router{
		drChan:  common.NewChanx[*schema.DataItem](),
		apiChan: common.NewChanx[*schema.DataItem](),
		apiConf: apiConf,
	}
}

// 数据入口
func dataEntry() {
	for item := range schema.ItemDrChan.Out {
		item := item
		ItemTotal.Inc()
		dr, ok := dataRouters.Load(item.APIName)
		if !ok {
			sampler.Error().Str("apiname", item.APIName).Int("len", len(item.APIName)).Msg("nonexistence")
			item.Release()
			continue
		}
		dr.drChan.In <- item
	}
	alarm.Error().Msg("Exception: DataRouter entry worker exited")
}
