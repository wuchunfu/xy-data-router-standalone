package datarouter

import (
	"github.com/fufuok/chanx"
	"github.com/fufuok/utils/xsync"
	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// JS 数据必须为 JSON 字典, 数据至少包含 1 个键值对, 即最小长为 7, 如: {"L":7}
const jsonMinLen = 7

var (
	// ES 数据分隔符
	esBodySep = []byte(conf.ESBodySep)

	// 以接口名为键的数据通道
	dataRouters = xsync.NewMap()

	// ItemTotal 收到的数据项计数
	ItemTotal xsync.Counter

	// ESChan ES 数据接收信道
	ESChan *chanx.UnboundedChan

	// DataProcessorTodoCount 待处理的数据项计数
	DataProcessorTodoCount xsync.Counter

	// DataProcessorDiscards 数据处理丢弃计数, 超过 ProcessorMaxWorkerSize
	DataProcessorDiscards xsync.Counter

	// ESDataItemDiscards 数据丢弃计数, 繁忙时丢弃可选接口的数据, 不写 ES
	ESDataItemDiscards xsync.Counter

	// ESDataTotal ES 收到数据数量计数
	ESDataTotal xsync.Counter

	// ESBulkCount ES Bulk 批量写入完成计数
	ESBulkCount xsync.Counter

	// ESBulkErrors ES Bulk 写入错误次数
	ESBulkErrors xsync.Counter

	// ESBulkTodoCount ES Bulk 待处理项计数
	ESBulkTodoCount xsync.Counter

	// ESBulkDiscards ES Bulk 写入丢弃协程数, 超过 ESBulkMaxWorkerSize
	ESBulkDiscards xsync.Counter

	// UDPRequestCount UDP 请求计数
	UDPRequestCount xsync.Counter

	// DataProcessorPool 数据处理协程池
	DataProcessorPool *ants.PoolWithFunc

	// ESBulkPool ES 写入协程池
	ESBulkPool *ants.PoolWithFunc
)

// 数据分发
type tDataRouter struct {
	// 数据接收信道
	drChan *chanx.UnboundedChan

	// 接口配置
	apiConf *conf.TAPIConf

	// 数据分发信道索引
	drOut *tDataRouterOut
}

// 数据分发信道
type tDataRouterOut struct {
	esChan  *chanx.UnboundedChan
	apiChan *chanx.UnboundedChan
}

// 数据处理
type tDataProcessor struct {
	dr   *tDataRouter
	data *schema.DataItem
}

// InitMain 程序启动时初始化
func InitMain() {
	// 初始化 ES 数据信道
	ESChan = common.NewChanx()

	// 初始化数据分发器
	initDataRouter()

	// 开启 ES 写入
	go esWorker()

	// 启动 UDP 接口服务
	go initUDPServer()

	// ES 索引头信息更新
	go updateESBulkHeader()

	// 初始化数据处理
	go initDataProcessorPool()
	go initESOptionalWrite()
	go initESBulkPool()
	go dataEntry()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	// 同步数据分发器配置
	initDataRouter()

	// 调节协程池
	tuneDataProcessorSize(conf.Config.DataConf.ProcessorSize)
	tuneESBulkWorkerSize(conf.Config.DataConf.ESBulkWorkerSize)
}

func Stop() {
	poolRelease()
}

func poolRelease() {
	DataProcessorPool.Release()
	ESBulkPool.Release()
}

// tuneDataProcessorSize 调节协程并发数
func tuneDataProcessorSize(n int) {
	DataProcessorPool.Tune(n)
}

// tuneESBulkWorkerSize 调节协程并发数
func tuneESBulkWorkerSize(n int) {
	ESBulkPool.Tune(n)
}

// 新数据信道
func newDataRouter(apiConf *conf.TAPIConf) *tDataRouter {
	return &tDataRouter{
		drChan:  common.NewChanx(),
		apiConf: apiConf,
		drOut: &tDataRouterOut{
			esChan:  ESChan,
			apiChan: common.NewChanx(),
		},
	}
}

// 数据入口
func dataEntry() {
	for item := range schema.ItemDrChan.Out {
		ItemTotal.Inc()
		item := item.(*schema.DataItem)
		dr, ok := dataRouters.Load(item.APIName)
		if !ok {
			common.LogSampled.Error().Str("apiname", item.APIName).Int("len", len(item.APIName)).Msg("nonexistence")
			item.Release()
			continue
		}
		dr.(*tDataRouter).drChan.In <- item
	}
}
