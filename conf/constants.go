package conf

import (
	"time"
)

const (
	APPName     = "XY.DataRouter"
	ProjectName = "xydatarouter"

	// LogLevel 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel = 3
	// LogSamplePeriodDur 抽样日志设置 (每秒最多 3 个日志)
	LogSamplePeriodDur = time.Second
	LogSampleBurst     = 3
	// LogFileMaxSize 每 100M 自动切割, 保留 30 天内最近 10 个日志文件
	LogFileMaxSize    = 100
	LogFileMaxBackups = 10
	LogFileMaxAge     = 30

	// WebServerAddr HTTP 接口端口
	WebServerAddr = ":6600"
	// TunServerAddr Tunnel 绑定端口
	TunServerAddr = ":6633"
	// WebSlowRespDuration Web 慢响应日志时间设置, 默认: > 1秒则记录
	WebSlowRespDuration = time.Second
	// WebErrorCodeLog HTTP 响应码日志记录, 默认: 500, 即大于等于 500 的状态码记录日志
	WebErrorCodeLog = 500
	// BodyLimit POST 最大 500M, Request Entity Too Large
	BodyLimit = 500 << 20
	// ESSlowQueryDuration ES 慢查询日志时间设置, 默认: > 10秒则记录
	ESSlowQueryDuration = 10 * time.Second
	// TunDialTimeout Tunnel 连接超时时间和数据发送队列容量
	TunDialTimeout   = 3 * time.Second
	TunSendQueueSize = 8192

	// UDPServerRAddr UDP 接口端口, 不应答(Echo包除外)
	UDPServerRAddr = ":6611"
	// UDPServerRWAddr UDP 接口端口, 每个包都应答字符: 1
	UDPServerRWAddr = ":6622"
	// UDPMaxRW 1. 在链路层, 由以太网的物理特性决定了数据帧的长度为 (46+18) - (1500+18)
	//    其中的 18 是数据帧的头和尾, 也就是说数据帧的内容最大为 1500 (不包括帧头和帧尾)
	//    即 MTU (Maximum Transmission Unit) 为 1500
	// 2. 在网络层, 因为 IP 包的首部要占用 20 字节, 所以这的 MTU 为 1500 - 20 = 1480
	// 3. 在传输层, 对于 UDP 包的首部要占用 8 字节, 所以这的 MTU 为 1480 - 8 = 1472
	// 4. UDP 协议中有 16 位的 UDP 报文长度, 即 UDP 报文长度不能超过 65536, 则数据最大为 65507
	UDPMaxRW = 65507
	// UDPGoReadNum1CPU UDP Goroutine 并发读取的数量 / CPU
	UDPGoReadNum1CPU = 50
	UDPGoReadNumMax  = 1000

	// ESBodySep ES 数据分隔符
	ESBodySep = "=-:-="
	// ESPostBatchNum ES 单次批量写入最大条数或最大字节数, 最大写入时间间隔
	ESPostBatchNum    = 4500
	ESPostBatchBytes  = 5 << 20
	ESPostMaxInterval = 500 * time.Millisecond
	// ESBulkWorkerSize ES 批量写入并发协程数, 最大排队数
	ESBulkWorkerSize    = 30
	ESBulkMaxWorkerSize = 800

	// DataChanSize 无限缓冲信道默认初始化缓冲大小
	DataChanSize = 50
	// DataChanMaxBufCap 无限缓冲信道最大缓冲数量, 0 为无限, 超过限制(DataChanSize + DataChanMaxBufCap)丢弃数据
	DataChanMaxBufCap = 0

	// DataProcessorSize 数据处理并发协程数
	DataProcessorSize          = 3000
	DataProcessorMaxWorkerSize = 100000

	// BaseSecretKeyName 项目基础密钥 (环境变量名)
	BaseSecretKeyName = "DR_BASE_SECRET_KEY"
	// BaseSecretSalt 用于解密基础密钥值的密钥 (编译在程序中)
	BaseSecretSalt = "Fufu@dr.777"

	// WatcherInterval 文件变化监控时间间隔(分)
	WatcherInterval = 1

	// HeartbeatIndex 心跳日志索引
	HeartbeatIndex = "monitor_heartbeat_report"
)
