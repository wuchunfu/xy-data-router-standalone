package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/json"
)

// 接口配置
type TJSONConf struct {
	SYSConf     TSYSConf   `json:"sys_conf"`
	APIConf     []TAPIConf `json:"api_conf"`
	ESWhiteList []string   `json:"es_white_list"`
	ESBlackList []string   `json:"es_black_list"`
}

// 主配置, 变量意义见配置文件中的描述及 constants.go 中的默认值
type TSYSConf struct {
	Debug                     bool       `json:"debug"`
	Log                       tLogConf   `json:"log"`
	PProfAddr                 string     `json:"pprof_addr"`
	WebServerAddr             string     `json:"web_server_addr"`
	EnableKeepalive           bool       `json:"enable_keepalive"`
	ReduceMemoryUsage         bool       `json:"reduce_memory_usage"`
	LimitBody                 int        `json:"limit_body"`
	LimitExpiration           int        `json:"limit_expiration"`
	LimitRequest              int        `json:"limit_request"`
	WebSlowResponse           int        `json:"web_slow_response"`
	WebErrCodeLog             int        `json:"web_errcode_log"`
	UDPServerRAddr            string     `json:"udp_server_raddr"`
	UDPServerRWAddr           string     `json:"udp_server_rwaddr"`
	UDPGoReadNum1CPU          int        `json:"udp_go_read_num_1cpu"`
	UDPProto                  string     `json:"udp_proto"`
	ESAddress                 []string   `json:"es_address"`
	ESPostBatchNum            int        `json:"es_post_batch_num"`
	ESPostBatchBytes          int        `json:"es_post_batch_mb"`
	ESPostMaxInterval         int        `json:"es_post_max_interval"`
	ESEnableRetry             bool       `json:"es_enable_retry"`
	ESDisableWrite            bool       `json:"es_disable_write"`
	ESSlowQuery               int        `json:"es_slow_query"`
	ESReentryCodes            []int      `json:"es_reentry_codes"`
	DataChanSize              int        `json:"data_chan_size"`
	DataProcessorSize         int        `json:"data_processor_size"`
	ESBulkWorkerSize          int        `json:"es_bulk_worker_size"`
	ESBulkMaxWorkerSize       int        `json:"es_bulk_max_worker_size"`
	MainConfig                TFilesConf `json:"main_config"`
	RestartMain               bool       `json:"restart_main"`
	WatcherInterval           int        `json:"watcher_interval"`
	HeartbeatIndex            string     `json:"heartbeat_index"`
	BaseSecretValue           string
	WebSlowRespDuration       time.Duration
	ESSlowQueryDuration       time.Duration
	ESPostMaxIntervalDuration time.Duration
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
	PeriodDur   time.Duration
}

type TAPIConf struct {
	APIName       string       `json:"api_name"`
	ESIndex       string       `json:"es_index"`
	ESIndexSplit  string       `json:"es_index_split"`
	RequiredField []string     `json:"required_field"`
	PostAPI       TPostAPIConf `json:"post_api"`
	ESBulkHeader  []byte
}

type TPostAPIConf struct {
	API      []string `json:"api"`
	Interval int      `json:"interval"`
}

type TFilesConf struct {
	Path            string `json:"path"`
	Method          string `json:"method"`
	SecretName      string `json:"secret_name"`
	API             string `json:"api"`
	Interval        int    `json:"interval"`
	SecretValue     string
	GetConfDuration time.Duration
	ConfigMD5       string
	ConfigVer       time.Time
}

func init() {
	confFile := flag.String("c", ConfigFile, "配置文件绝对路径")
	flag.Parse()
	ConfigFile = *confFile
	if err := LoadConf(); err != nil {
		log.Fatalln("Failed to initialize config:", err, "\nbye.")
	}
}

// 加载配置
func LoadConf() error {
	config, apiConfig, whiteList, blackList, err := readConf()
	if err != nil {
		return err
	}

	Config = *config
	APIConfig = apiConfig
	ESWhiteListConfig = whiteList
	ESBlackListConfig = blackList

	return nil
}

// 读取配置
func readConf() (*TJSONConf, map[string]*TAPIConf, map[*net.IPNet]struct{}, map[*net.IPNet]struct{}, error) {
	body, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var config *TJSONConf
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, nil, nil, nil, err
	}

	// 基础密钥 Key
	config.SYSConf.BaseSecretValue = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)
	if config.SYSConf.BaseSecretValue == "" {
		return nil, nil, nil, nil, fmt.Errorf("%s cannot be empty", BaseSecretKeyName)
	}

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	if config.SYSConf.Log.Level > 7 || config.SYSConf.Log.Level < -1 {
		config.SYSConf.Log.Level = LogLevel
	}

	// ES 批量写入错误日志
	if config.SYSConf.Log.ESBulkLevel > 7 || config.SYSConf.Log.ESBulkLevel < -1 {
		config.SYSConf.Log.ESBulkLevel = LogLevel
	}

	// 抽样日志设置 (x 秒 n 条)
	if config.SYSConf.Log.Burst < 0 || config.SYSConf.Log.Period < 0 {
		config.SYSConf.Log.PeriodDur = LogSamplePeriodDur
		config.SYSConf.Log.Burst = LogSampleBurst
	} else {
		config.SYSConf.Log.PeriodDur = time.Duration(config.SYSConf.Log.Period) * time.Second
	}

	// 日志文件
	if config.SYSConf.Log.File == "" {
		config.SYSConf.Log.File = LogFile
	}

	// 日志大小和保存设置
	if config.SYSConf.Log.MaxSize < 1 {
		config.SYSConf.Log.MaxSize = LogFileMaxSize
	}
	if config.SYSConf.Log.MaxBackups < 1 {
		config.SYSConf.Log.MaxBackups = LogFileMaxBackups
	}
	if config.SYSConf.Log.MaxAge < 1 {
		config.SYSConf.Log.MaxAge = LogFileMaxAge
	}

	// 单个 CPU 的 UDP 并发读取协程数, 默认为 50
	if config.SYSConf.UDPGoReadNum1CPU < 10 {
		config.SYSConf.UDPGoReadNum1CPU = UDPGoReadNum1CPU
	}

	// 数据分发通道缓存大小
	if config.SYSConf.DataChanSize < 1 {
		config.SYSConf.DataChanSize = DataChanSize
	}

	// 优先使用配置中的绑定参数(HTTP)
	if config.SYSConf.WebServerAddr == "" {
		config.SYSConf.WebServerAddr = WebServerAddr
	}

	// 优先使用配置中的绑定参数(UDP带应答)
	if config.SYSConf.UDPServerRWAddr == "" {
		config.SYSConf.UDPServerRWAddr = UDPServerRWAddr
	}

	// 优先使用配置中的绑定参数(UDP不带应答)
	if config.SYSConf.UDPServerRAddr == "" {
		config.SYSConf.UDPServerRAddr = UDPServerRAddr
	}

	// ES 慢查询日志时间设置
	if config.SYSConf.ESSlowQuery < 1 {
		config.SYSConf.ESSlowQueryDuration = ESSlowQueryDuration
	} else {
		config.SYSConf.ESSlowQueryDuration = time.Duration(config.SYSConf.ESSlowQuery) * time.Second
	}

	// Web 慢响应日志时间设置
	if config.SYSConf.WebSlowResponse < 1 {
		config.SYSConf.WebSlowRespDuration = WebSlowRespDuration
	} else {
		config.SYSConf.WebSlowRespDuration = time.Duration(config.SYSConf.WebSlowResponse) * time.Second
	}

	// HTTP 响应码日志设置, 默认 >= 500
	if config.SYSConf.WebErrCodeLog < 1 {
		config.SYSConf.WebErrCodeLog = WebErrorCodeLog
	}

	// 以接口名为键的配置集合
	apiConfig := make(map[string]*TAPIConf)
	for _, cfg := range config.APIConf {
		apiConf := cfg
		apiname := strings.TrimSpace(apiConf.APIName)
		if apiname == "" {
			continue
		}
		apiConfig[apiConf.APIName] = &apiConf
	}

	// ES 查询接口 IP 白名单
	whiteList, err := getIPNetList(config.ESWhiteList)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// ES 查询接口 IP 白名单
	blackList, err := getIPNetList(config.ESBlackList)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// 每次获取远程主配置的时间间隔, < 30 秒则禁用该功能
	if config.SYSConf.MainConfig.Interval > 29 {
		// 远程获取主配置 API, 解密 SecretName
		if config.SYSConf.MainConfig.SecretName != "" {
			config.SYSConf.MainConfig.SecretValue = utils.GetenvDecrypt(config.SYSConf.MainConfig.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.SYSConf.MainConfig.SecretValue == "" {
				return nil, nil, nil, nil, fmt.Errorf("%s cannot be empty", config.SYSConf.MainConfig.SecretName)
			}
		}
		config.SYSConf.MainConfig.GetConfDuration = time.Duration(config.SYSConf.MainConfig.Interval) * time.Second
	}
	config.SYSConf.MainConfig.Path = ConfigFile

	// 文件变化监控时间间隔
	if config.SYSConf.WatcherInterval < 1 {
		config.SYSConf.WatcherInterval = WatcherInterval
	}

	// 心跳日志索引
	if config.SYSConf.HeartbeatIndex == "" {
		config.SYSConf.HeartbeatIndex = HeartbeatIndex
	}

	// HTTP 请求体限制, -1 表示无限
	if config.SYSConf.LimitBody == 0 {
		config.SYSConf.LimitBody = BodyLimit
	}

	// 数据处理并发协程数限制
	if config.SYSConf.DataProcessorSize < 10 {
		config.SYSConf.DataProcessorSize = DataProcessorSize
	}

	// ES Bulk 写入并发协程限制
	if config.SYSConf.ESBulkWorkerSize < 1 {
		config.SYSConf.ESBulkWorkerSize = ESBulkWorkerSize
	}
	if config.SYSConf.ESBulkMaxWorkerSize < 100 {
		config.SYSConf.ESBulkMaxWorkerSize = ESBulkMaxWorkerSize
	}

	// ES Bulk 单次批量数量
	if config.SYSConf.ESPostBatchNum < 100 {
		config.SYSConf.ESPostBatchNum = ESPostBatchNum
	}

	// ES Bulk 单次批量大小
	if config.SYSConf.ESPostBatchBytes < 1 {
		config.SYSConf.ESPostBatchBytes = ESPostBatchBytes
	} else {
		// 配置文件单位是 MB
		config.SYSConf.ESPostBatchBytes = config.SYSConf.ESPostBatchBytes << 20
	}

	// ES Bulk 单次批量最大时间间隔
	if config.SYSConf.ESPostMaxInterval < 100 {
		config.SYSConf.ESPostMaxIntervalDuration = ESPostMaxInterval
	} else {
		config.SYSConf.ESPostMaxIntervalDuration = time.Duration(config.SYSConf.ESPostMaxInterval) * time.Millisecond
	}

	return config, apiConfig, whiteList, blackList, nil
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
