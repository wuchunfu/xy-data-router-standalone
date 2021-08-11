package conf

import (
	"net"
	"path/filepath"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

// RootPath 运行绝对路径
var RootPath = utils.ExecutableDir(true)

// FilePath 配置文件绝对路径
var FilePath = filepath.Join(RootPath, "..", "etc")

// ConfigFile 默认配置文件路径
var ConfigFile = filepath.Join(FilePath, ProjectName+".json")

// LogDir 日志路径
var LogDir = filepath.Join(RootPath, "..", "log")
var LogFile = filepath.Join(LogDir, ProjectName+".log")

// LogDaemon 守护日志
var LogDaemon = filepath.Join(LogDir, "daemon.log")

// Config 所有配置
var Config TJSONConf

// APIConfig 以接口名为键的配置
var APIConfig map[string]*TAPIConf

// ESWhiteListConfig ES 查询接口 IP 白名单配置
var ESWhiteListConfig map[*net.IPNet]struct{}

// ESBlackListConfig ES 上报接口 IP 黑名单配置
var ESBlackListConfig map[*net.IPNet]struct{}

// UDPESIndexField UDP 接口 ES 索引字段
var UDPESIndexField = "_x"

// ReqUserAgent 请求名称
var ReqUserAgent = req.Header{"User-Agent": WebAPPName + "/" + CurrentVersion}

// ForwardTunnel 上联地址, 指定后将启动客户端, 将所有数据转交到 Tunnel Server
var ForwardTunnel = ""
