package service

import (
	"github.com/fufuok/chanx"
	"github.com/fufuok/utils/xsync"
	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

// JS 数据必须为 JSON 字典, 数据至少包含 1 个键值对, 即最小长为 7, 如: {"L":7}
const jsonMinLen = 7

var (
	// ES 数据分隔符
	esBodySep = []byte(conf.ESBodySep)

	// 以接口名为键的数据通道
	dataRouters = xsync.NewMap()

	// ES 数据接收信道
	esChan *chanx.UnboundedChan

	// TunChan Tun 数据信道
	TunChan *chanx.UnboundedChan

	// 计数开始时间
	counterStartTime = common.GetGlobalTime()

	// 待处理的数据项计数
	dataProcessorTodoCount xsync.Counter

	// 数据处理丢弃计数, 超过 ProcessorMaxWorkerSize
	dataProcessorDiscards xsync.Counter

	// 数据丢弃计数, 繁忙时丢弃可选接口的数据, 不写 ES
	esDataItemDiscards xsync.Counter

	// ES 收到数据数量计数
	esDataTotal xsync.Counter

	// ES Bulk 批量写入完成计数
	esBulkCount xsync.Counter

	// ES Bulk 写入错误次数
	esBulkErrors xsync.Counter

	// ES Bulk 待处理项计数
	esBulkTodoCount xsync.Counter

	// ES Bulk 写入丢弃协程数, 超过 ESBulkMaxWorkerSize
	esBulkDiscards xsync.Counter

	// HTTPRequestCount HTTP 请求计数
	HTTPRequestCount    xsync.Counter
	HTTPBadRequestCount xsync.Counter

	// TunRecvCount Tunnel 服务端接收和客户端发送计数
	TunRecvCount     xsync.Counter
	TunRecvBadCount  xsync.Counter
	TunSendCount     xsync.Counter
	TunSendErrors    xsync.Counter
	TunDataTotal     xsync.Counter
	TunCompressTotal xsync.Counter

	// UDP 请求计数
	udpRequestCount xsync.Counter

	// 数据处理协程池
	dataProcessorPool *ants.PoolWithFunc

	// ES 写入协程池
	esBulkPool *ants.PoolWithFunc
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

func InitService() {
	// 初始化 ES 数据信道
	esChan = newChanx()

	// 初始化 Tun 数据信道
	TunChan = newChanx()

	// 开启 ES 写入
	go esWorker()

	// 初始化数据分发器
	InitDataRouter()

	// 启动 UDP 接口服务
	go initUDPServer()

	// 心跳服务
	go initHeartbeat()

	// ES 索引头信息更新
	go updateESBulkHeader()

	// 初始化数据处理
	go initDataProcessorPool()
	go initESOptionalWrite()
	go initESBulkPool()

	// 初始化运行时参数
	go initRuntime()
}

func Stop() {
	poolRelease()
}

func poolRelease() {
	dataProcessorPool.Release()
	esBulkPool.Release()
}

// TuneDataProcessorSize 调节协程并发数
func TuneDataProcessorSize(n int) {
	dataProcessorPool.Tune(n)
}

// TuneESBulkWorkerSize 调节协程并发数
func TuneESBulkWorkerSize(n int) {
	esBulkPool.Tune(n)
}

// 初始化无限缓冲信道
func newChanx() *chanx.UnboundedChan {
	return chanx.NewUnboundedChan(conf.Config.DataConf.ChanSize, conf.Config.DataConf.ChanMaxBufCap)
}

// 新数据信道
func newDataRouter(apiConf *conf.TAPIConf) *tDataRouter {
	return &tDataRouter{
		drChan:  newChanx(),
		apiConf: apiConf,
		drOut: &tDataRouterOut{
			esChan:  esChan,
			apiChan: newChanx(),
		},
	}
}
