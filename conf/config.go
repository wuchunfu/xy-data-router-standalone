package conf

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/internal/json"
)

// tJSONConf 接口配置
type tJSONConf struct {
	SYSConf     tSYSConf   `json:"sys_conf"`
	MainConf    TFilesConf `json:"main_conf"`
	LogConf     tLogConf   `json:"log_conf"`
	WebConf     tWebConf   `json:"web_conf"`
	UDPConf     tUDPConf   `json:"udp_conf"`
	TunConf     tTunConf   `json:"tun_conf"`
	DataConf    tDataConf  `json:"data_conf"`
	APIConf     []TAPIConf `json:"api_conf"`
	ESWhiteList []string   `json:"es_white_list"`
	ESBlackList []string   `json:"es_black_list"`
	StateConf   tStateConf
}

// tSYSConf 主配置, 变量意义见配置文件中的描述及 constants.go 中的默认值
type tSYSConf struct {
	Debug                   bool   `json:"debug"`
	PProfAddr               string `json:"pprof_addr"`
	RestartMain             bool   `json:"restart_main"`
	WatcherInterval         int    `json:"watcher_interval"`
	ReqTimeout              int    `json:"req_timeout"`
	ReqMaxRetries           int    `json:"req_max_retries"`
	ReqTimeoutDuration      time.Duration
	WatcherIntervalDuration time.Duration
	BaseSecretValue         string
}

type tLogConf struct {
	Level       int    `json:"level"`
	NoColor     bool   `json:"no_color"`
	File        string `json:"file"`
	Period      int    `json:"period"`
	Burst       uint32 `json:"burst"`
	MaxSize     int    `json:"max_size"`
	MaxBackups  int    `json:"max_backups"`
	MaxAge      int    `json:"max_age"`
	ESBulkLevel int    `json:"es_bulk_level"`
	ESIndex     string `json:"es_index"`
	PeriodDur   time.Duration
}

type tWebConf struct {
	ServerAddr              string   `json:"server_addr"`
	ServerHttpsAddr         string   `json:"server_https_addr"`
	DisableKeepalive        bool     `json:"disable_keepalive"`
	ReduceMemoryUsage       bool     `json:"reduce_memory_usage"`
	LimitBody               int      `json:"limit_body"`
	LimitExpiration         int      `json:"limit_expiration"`
	LimitRequest            int      `json:"limit_request"`
	SlowResponse            int      `json:"slow_response"`
	ErrCodeLog              int      `json:"errcode_log"`
	ProxyHeader             string   `json:"proxy_header"`
	EnableTrustedProxyCheck bool     `json:"enable_trusted_proxy_check"`
	TrustedProxies          []string `json:"trusted_proxies"`
	ESAPITimeoutSecond      int      `json:"esapi_timeout_second"`
	ESSlowQuery             int      `json:"es_slow_query"`
	ESSlowQueryDuration     time.Duration
	SlowResponseDuration    time.Duration
	ESAPITimeout            time.Duration
	CertFile                string
	KeyFile                 string
}

type tUDPConf struct {
	ServerRAddr   string `json:"server_raddr"`
	ServerRWAddr  string `json:"server_rwaddr"`
	GoReadNum1CPU int    `json:"go_read_num_1cpu"`
	Proto         string `json:"proto"`
	GoReadNum     int
}

type tTunConf struct {
	ServerAddr      string `json:"server_addr"`
	ClientNum1CPU   int    `json:"client_num_1_cpu"`
	SendQueueSize   int    `json:"send_queue_size"`
	CompressMinSize uint64 `json:"compress_min_size"`
	ClientNum       int
}

type tDataConf struct {
	ESAddress                 []string `json:"es_address"`
	ESDisableWrite            bool     `json:"es_disable_write"`
	ESPostBatchNum            int      `json:"es_post_batch_num"`
	ESPostBatchMB             int      `json:"es_post_batch_mb"`
	ESPostMaxInterval         int      `json:"es_post_max_interval"`
	ESRetryOnStatus           []int    `json:"es_retry_on_status"`
	ESMaxRetries              int      `json:"es_max_retries"`
	ESDisableRetry            bool     `json:"es_disable_retry"`
	ESBulkWorkerSize          int      `json:"es_bulk_worker_size"`
	ESBulkMaxWorkerSize       int      `json:"es_bulk_max_worker_size"`
	ESBusyPercent             float64  `json:"es_busy_percent"`
	ChanSize                  int      `json:"chan_size"`
	ChanMaxBufCap             int      `json:"chan_max_buf_cap"`
	ProcessorSize             int      `json:"processor_size"`
	ProcessorMaxWorkerSize    int      `json:"processor_max_worker_size"`
	ESPostBatchBytes          int
	ESPostMaxIntervalDuration time.Duration
}

type tStateConf struct {
	CheckESBulkResult bool
	CheckESBulkErrors bool
	RecordESBulkBody  bool
}

type TAPIConf struct {
	APIName            string       `json:"api_name"`
	ESDisableWrite     bool         `json:"es_disable_write"`
	ESOptionalWrite    bool         `json:"es_optional_write"`
	ESIndex            string       `json:"es_index"`
	ESIndexSplit       string       `json:"es_index_split"`
	ESPipeline         string       `json:"es_pipeline"`
	RequiredField      []string     `json:"required_field"`
	PostAPI            TPostAPIConf `json:"post_api"`
	ESBulkHeader       []byte
	ESBulkHeaderLength int
}

type TPostAPIConf struct {
	API        []string `json:"api"`
	Interval   int      `json:"interval"`
	BatchNum   int      `json:"batch_num"`
	BatchMB    int      `json:"batch_mb"`
	BatchBytes int
}

type TFilesConf struct {
	Path            string `json:"-"`
	Method          string `json:"method"`
	SecretName      string `json:"secret_name"`
	API             string `json:"api"`
	Interval        int    `json:"interval"`
	SecretValue     string
	GetConfDuration time.Duration
}

// LoadConf 加载配置
func LoadConf() error {
	config, apiConfig, whiteList, blackList, err := readConf()
	if err != nil {
		return err
	}

	Debug = config.SYSConf.Debug
	Config = config
	APIConfig = apiConfig
	ESWhiteListConfig = whiteList
	ESBlackListConfig = blackList

	return nil
}

// 读取配置
func readConf() (
	config *tJSONConf,
	apiConfig map[string]*TAPIConf,
	whiteList map[*net.IPNet]struct{},
	blackList map[*net.IPNet]struct{},
	err error,
) {
	var body []byte
	body, err = os.ReadFile(ConfigFile)
	if err != nil {
		return
	}

	config = new(tJSONConf)
	if err = json.Unmarshal(body, config); err != nil {
		return
	}

	// 基础密钥 Key
	config.SYSConf.BaseSecretValue = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)
	if config.SYSConf.BaseSecretValue == "" {
		err = fmt.Errorf("%s cannot be empty", BaseSecretKeyName)
		return
	}

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	if config.LogConf.Level > 7 || config.LogConf.Level < -1 {
		config.LogConf.Level = LogLevel
	}

	// ES 批量写入错误日志
	if config.LogConf.ESBulkLevel > 7 || config.LogConf.ESBulkLevel < -1 {
		config.LogConf.ESBulkLevel = LogLevel
	}

	// 抽样日志设置 (x 秒 n 条)
	if config.LogConf.Burst < 0 || config.LogConf.Period < 0 {
		config.LogConf.PeriodDur = LogSamplePeriodDur
		config.LogConf.Burst = LogSampleBurst
	} else {
		config.LogConf.PeriodDur = time.Duration(config.LogConf.Period) * time.Second
	}

	// 日志文件
	if config.LogConf.File == "" {
		config.LogConf.File = LogFile
	}

	// 日志大小和保存设置
	if config.LogConf.MaxSize < 1 {
		config.LogConf.MaxSize = LogFileMaxSize
	}
	if config.LogConf.MaxBackups < 1 {
		config.LogConf.MaxBackups = LogFileMaxBackups
	}
	if config.LogConf.MaxAge < 1 {
		config.LogConf.MaxAge = LogFileMaxAge
	}

	// 单个 CPU 的 UDP 并发读取协程数, 默认为 2
	if config.UDPConf.GoReadNum1CPU < 1 {
		config.UDPConf.GoReadNum1CPU = UDPGoReadNum1CPU
	}
	config.UDPConf.GoReadNum = utils.MinInt(config.UDPConf.GoReadNum1CPU*runtime.NumCPU(), UDPGoReadNumMax)

	// UDP 协议原型
	if config.UDPConf.Proto != "gnet" {
		config.UDPConf.Proto = "default"
	}

	// 数据分发通道缓存大小
	if config.DataConf.ChanSize < 1 {
		config.DataConf.ChanSize = DataChanSize
	}

	// 数据分发通道最大缓存数限制, 0 为无限
	if config.DataConf.ChanMaxBufCap < 0 {
		config.DataConf.ChanMaxBufCap = DataChanMaxBufCap
	}

	// 优先使用配置中的绑定参数(HTTP)
	if config.WebConf.ServerAddr == "" {
		config.WebConf.ServerAddr = WebServerAddr
	}

	// 证书文件存在时开启 HTTPS
	config.WebConf.CertFile = os.Getenv(WebCertFileEnv)
	config.WebConf.KeyFile = os.Getenv(WebKeyFileEnv)
	if utils.IsFile(config.WebConf.CertFile) && utils.IsFile(config.WebConf.KeyFile) {
		// 优先使用配置中的绑定参数(HTTPS)
		if config.WebConf.ServerHttpsAddr == "" {
			config.WebConf.ServerHttpsAddr = WebServerHttpsAddr
		}
	} else {
		config.WebConf.ServerHttpsAddr = ""
	}

	// 优先使用配置中的绑定参数(Tunnel)
	if config.TunConf.ServerAddr == "" {
		config.TunConf.ServerAddr = TunServerAddr
	}

	// 优先使用配置中的绑定参数(UDP带应答)
	if config.UDPConf.ServerRWAddr == "" {
		config.UDPConf.ServerRWAddr = UDPServerRWAddr
	}

	// 优先使用配置中的绑定参数(UDP不带应答)
	if config.UDPConf.ServerRAddr == "" {
		config.UDPConf.ServerRAddr = UDPServerRAddr
	}

	// Tunnel 发送队列容量
	if config.TunConf.SendQueueSize < TunSendQueueSize {
		config.TunConf.SendQueueSize = TunSendQueueSize
	}

	// Tunnel 压缩传输数据最小字节数
	if config.TunConf.CompressMinSize < TunCompressMinSize {
		config.TunConf.CompressMinSize = TunCompressMinSize
	}

	// Tunnel 单个 CPU 的发送客户端数, 默认为 2
	if config.TunConf.ClientNum1CPU < 1 {
		config.TunConf.ClientNum1CPU = TunClientNum1CPU
	}
	config.TunConf.ClientNum = utils.MinInt(config.TunConf.ClientNum1CPU*runtime.NumCPU(), TunClientNumMax)

	// ES 慢查询日志时间设置
	if config.WebConf.ESSlowQuery < 1 {
		config.WebConf.ESSlowQueryDuration = ESSlowQueryDuration
	} else {
		config.WebConf.ESSlowQueryDuration = time.Duration(config.WebConf.ESSlowQuery) * time.Second
	}

	// Web 慢响应日志时间设置
	if config.WebConf.SlowResponse < 1 {
		config.WebConf.SlowResponseDuration = WebSlowResponseDuration
	} else {
		config.WebConf.SlowResponseDuration = time.Duration(config.WebConf.SlowResponse) * time.Second
	}

	// HTTP 响应码日志设置, 默认 >= 500
	if config.WebConf.ErrCodeLog < 1 {
		config.WebConf.ErrCodeLog = WebErrorCodeLog
	}

	// ES 查询请求代理时的超时秒数, 默认: 30s
	if config.WebConf.ESAPITimeoutSecond < 1 {
		config.WebConf.ESAPITimeout = WebESAPITimeout
	} else {
		config.WebConf.ESAPITimeout = time.Duration(config.WebConf.ESAPITimeoutSecond) * time.Second
	}

	// 以接口名为键的配置集合
	apiConfig = make(map[string]*TAPIConf)
	for _, cfg := range config.APIConf {
		apiConf := cfg
		apiname := strings.TrimSpace(apiConf.APIName)
		if apiname == "" {
			continue
		}

		apiConf.ESPipeline = strings.TrimSpace(apiConf.ESPipeline)
		if len(apiConf.PostAPI.API) > 0 && apiConf.PostAPI.Interval > 0 {
			// 单次汇聚最大数量
			if apiConf.PostAPI.BatchNum < 1 {
				apiConf.PostAPI.BatchNum = APIPostBatchNum
			}
			// 单次汇聚最大字节大小
			if apiConf.PostAPI.BatchMB < 1 {
				apiConf.PostAPI.BatchBytes = APIPostBatchBytes
			} else {
				// 配置文件单位是 MB
				apiConf.PostAPI.BatchBytes = apiConf.PostAPI.BatchMB << 20
			}
		} else {
			// 禁用该 API 数据分发功能
			apiConf.PostAPI.Interval = 0
		}

		apiConfig[apiname] = &apiConf
	}

	// ES 查询接口 IP 白名单
	whiteList, err = getIPNetList(config.ESWhiteList)
	if err != nil {
		return
	}

	// 接口访问 IP 黑名单
	blackList, err = getIPNetList(config.ESBlackList)
	if err != nil {
		return
	}

	// 每次获取远程主配置的时间间隔, < 30 秒则禁用该功能
	if config.MainConf.Interval > 29 {
		// 远程获取主配置 API, 解密 SecretName
		if config.MainConf.SecretName != "" {
			config.MainConf.SecretValue = utils.GetenvDecrypt(config.MainConf.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.MainConf.SecretValue == "" {
				err = fmt.Errorf("%s cannot be empty", config.MainConf.SecretName)
				return
			}
		}
		config.MainConf.GetConfDuration = time.Duration(config.MainConf.Interval) * time.Second
	}
	config.MainConf.Path = ConfigFile

	// 文件变化监控时间间隔
	if config.SYSConf.WatcherInterval < 1 {
		config.SYSConf.WatcherIntervalDuration = WatcherIntervalDuration
	} else {
		config.SYSConf.WatcherIntervalDuration = time.Duration(config.SYSConf.WatcherInterval) * time.Minute
	}

	// ES 查询接口日志索引
	if config.LogConf.ESIndex == "" {
		config.LogConf.ESIndex = LogESIndex
	}

	// HTTP 请求体限制, -1 表示无限
	if config.WebConf.LimitBody == 0 {
		config.WebConf.LimitBody = BodyLimit
	}

	// 数据处理并发协程数限制
	if config.DataConf.ProcessorSize < 10 {
		config.DataConf.ProcessorSize = DataProcessorSize
	}
	if config.DataConf.ProcessorMaxWorkerSize < 10000 {
		config.DataConf.ProcessorMaxWorkerSize = DataProcessorMaxWorkerSize
	}

	// ES Bulk 写入并发协程限制
	if config.DataConf.ESBulkWorkerSize < 1 {
		config.DataConf.ESBulkWorkerSize = ESBulkWorkerSize
	}
	if config.DataConf.ESBulkMaxWorkerSize < ESBulkMinWorkerSize {
		config.DataConf.ESBulkMaxWorkerSize = ESBulkMaxWorkerSize
	}

	// ES 基于排队数的繁忙比率定义
	if config.DataConf.ESBusyPercent < 0.01 || config.DataConf.ESBusyPercent > 1 {
		config.DataConf.ESBusyPercent = ESBusyPercent
	}

	// ES Bulk 单次批量数量
	if config.DataConf.ESPostBatchNum < 100 {
		config.DataConf.ESPostBatchNum = ESPostBatchNum
	}

	// ES Bulk 单次批量大小
	if config.DataConf.ESPostBatchMB < 1 {
		config.DataConf.ESPostBatchBytes = ESPostBatchBytes
	} else {
		// 配置文件单位是 MB
		config.DataConf.ESPostBatchBytes = config.DataConf.ESPostBatchMB << 20
	}

	// ES Bulk 单次批量最大时间间隔
	if config.DataConf.ESPostMaxInterval < 100 {
		config.DataConf.ESPostMaxIntervalDuration = ESPostMaxInterval
	} else {
		config.DataConf.ESPostMaxIntervalDuration = time.Duration(config.DataConf.ESPostMaxInterval) * time.Millisecond
	}

	// 作为客户端发起请求默认超时时间, 数据分发
	if config.SYSConf.ReqTimeout < 1 {
		config.SYSConf.ReqTimeoutDuration = ReqTimeoutDuration
	} else {
		config.SYSConf.ReqTimeoutDuration = time.Duration(config.SYSConf.ReqTimeout) * time.Second
	}

	// 更新状态类配置
	config.StateConf.CheckESBulkResult = config.LogConf.Level <= int(zerolog.WarnLevel)
	config.StateConf.CheckESBulkErrors = config.LogConf.ESBulkLevel <= int(zerolog.WarnLevel)
	config.StateConf.RecordESBulkBody = config.LogConf.ESBulkLevel <= int(zerolog.InfoLevel)

	ver := GetFilesVer(ConfigFile)
	ver.MD5 = utils.MD5BytesHex(body)
	ver.LastUpdate = time.Now()
	if config.SYSConf.Debug {
		fmt.Printf("\n\n%s\n\n", json.MustJSONIndent(config))
		fmt.Printf("\nWhitelist:\n%s\n\n", whiteList)
		fmt.Printf("\nBlackList:\n%s\n\n", blackList)
	}
	return
}

// IP 配置转换
func getIPNetList(ips []string) (map[*net.IPNet]struct{}, error) {
	ipNets := make(map[*net.IPNet]struct{})
	for _, ip := range ips {
		// 排除空白行, __ 或 # 开头的注释行
		ip = strings.TrimSpace(ip)
		if ip == "" || strings.HasPrefix(ip, "__") || strings.HasPrefix(ip, "#") {
			continue
		}
		// 补全掩码
		if !strings.Contains(ip, "/") {
			if strings.Contains(ip, ":") {
				ip = ip + "/128"
			} else {
				ip = ip + "/32"
			}
		}
		// 转为网段
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, err
		}
		ipNets[ipNet] = struct{}{}
	}

	return ipNets, nil
}
