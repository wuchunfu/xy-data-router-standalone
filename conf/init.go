package conf

import (
	"net"
	"path/filepath"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
)

// 运行绝对路径
var RootPath = utils.ExecutableDir(true)

// 配置文件绝对路径
var FilePath = filepath.Join(RootPath, "..", "etc")

// 默认配置文件路径
var ConfigFile = filepath.Join(FilePath, ProjectName+".json")

// 日志路径
var LogDir = filepath.Join(RootPath, "..", "log")
var LogFile = filepath.Join(LogDir, ProjectName+".log")

// 守护日志
var LogDaemon = filepath.Join(LogDir, "daemon.log")

// 所有配置
var Config TJSONConf

// 以接口名为键的配置
var APIConfig map[string]*TAPIConf

// ES 接口 IP 白名单配置
var ESWhiteListConfig map[*net.IPNet]struct{}

// UDP 接口 ES 索引字段
var UDPESIndexField = "_x"

// 请求名称
var ReqUserAgent = req.Header{"User-Agent": WebAPPName + "/" + CurrentVersion}
