package service

import (
	"github.com/fufuok/chanx"
	"github.com/fufuok/cmap"
	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// 数据项
type tDataItem struct {
	// 接口名称
	apiname string

	// 客户端 IP
	ip string

	// HTTP / UDP 数据
	body *[]byte
}

// 数据分发
type tDataRouter struct {
	// 数据接收信道
	drChan chanx.UnboundedChan

	// 接口配置
	apiConf *conf.TAPIConf

	// 数据分发信道索引
	drOut *tDataRouterOut
}

// 数据分发信道
type tDataRouterOut struct {
	esChan  chanx.UnboundedChan
	apiChan chanx.UnboundedChan
}

// 数据处理
type tDataProcessor struct {
	dr   *tDataRouter
	data *tDataItem
}

var (
	ln         = []byte("\n")
	jsArrLeft  = []byte("[")
	jsArrRight = []byte("]")

	// ES 数据分隔符
	esBodySep = []byte(conf.ESBodySep)

	// 以接口名为键的数据通道
	dataRouters = cmap.New()

	// ES 数据接收信道
	esChan chanx.UnboundedChan

	// WsHub 数据信道
	wsHubChan chanx.UnboundedChan

	// 计数开始时间
	counterStartTime = common.GetGlobalTime()

	// 待处理的数据项计数
	dataProcessorTodoCounters int64 = 0

	// 数据处理丢弃计数, 超过 DataProcessorMaxWorkerSize
	dataProcessorDropCounters uint64 = 0

	// ES 收到数据数量计数
	esDataCounters uint64 = 0

	// ES Bulk 批量写入完成计数
	esBulkDoneCounters uint64 = 0

	// ES Bulk 写入错误次数
	esBulkErrors uint64 = 0

	// ES Bulk 待处理项计数
	esBulkTodoCounters int64 = 0

	// ES Bulk 写入丢弃协程数, 超过 ESBulkMaxWorkerSize
	esBulkDropCounters uint64 = 0

	// HTTP 请求计数
	HTTPRequestCounters    uint64 = 0
	HTTPBadRequestCounters uint64 = 0

	// WsHub 请求计数
	WsHubRequestCounters uint64 = 0

	// UDP 请求计数
	UDPRequestCounters uint64 = 0

	// 数据处理协程池
	dataProcessorPool *ants.PoolWithFunc

	// ES 写入协程池
	esBulkPool *ants.PoolWithFunc
)

func InitService() {
	// 初始化 ES 数据信道
	esChan = newChanx()

	// 初始化 WsHub 数据信道
	wsHubChan = newChanx()

	// 开启 ES 写入
	go esWorker()

	// 初始化数据分发器
	InitDataRouter()

	// 启动 UDP 接口服务
	go initUDPServer()

	// 启动 WsHub 客户端
	go initWsHubClient()

	// 心跳服务
	go initHeartbeat()

	// ES 索引头信息更新
	go updateESBulkHeader()

	// 初始化协程池
	go initDataProcessorPool()
	go initESBulkPool()

	// 初始化运行时参数
	go initRuntime()
}

func PoolRelease() {
	dataProcessorPool.Release()
	esBulkPool.Release()
}

// 调节协程并发数
func TuneDataProcessorSize(n int) {
	dataProcessorPool.Tune(n)
}

// 调节协程并发数
func TuneESBulkWorkerSize(n int) {
	esBulkPool.Tune(n)
}

// 初始化无限缓冲信道
func newChanx() chanx.UnboundedChan {
	return chanx.NewUnboundedChan(conf.Config.SYSConf.DataChanSize)
}

// 新数据项
func newDataItem(apiname, ip string, body *[]byte) *tDataItem {
	return &tDataItem{apiname, ip, body}
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
