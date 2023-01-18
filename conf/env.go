package conf

import (
	"os"
)

// 加载环境变量
func loadEnvConf() {
	if ForwardHost == "" {
		ForwardHost = os.Getenv(ForwardHostEnv)
	}
}
