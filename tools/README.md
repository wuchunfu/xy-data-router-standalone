# 环境变量加密工具

用于项目中敏感配置项加解密. 比如各类 API secret

详见: https://github.com/fufuok/utils

1. 项目 git 中不会出现明文信息
2. 运行环境中也不会见到明文信息, 也不能通过环境变量值解密

## 1. 前置

1. 先这个项目用到的基础密钥名称设置到固定环境变量 `ENV_TOOLS_NAME`
2. 再把项目的基础加密密钥(明文, 也可以自定义密文, 见示例), 设置到环境变量, 名称即为项目中要用到的环境变量名称

```shell
# Linux
# 告知该加密工具, 项目的基础密钥环境变量名称是: XY_PROJECT_BASE_SECRET_KEY
export ENV_TOOLS_NAME=XY_PROJECT_BASE_SECRET_KEY
# 设置基础加密密钥(明文)到项目中用于获取密钥的环境变量名称: XY_PROJECT_BASE_SECRET_KEY
export XY_PROJECT_BASE_SECRET_KEY=myBASEkeyValue123
# Windows
set ENV_TOOLS_NAME=XY_PROJECT_BASE_SECRET_KEY
set XY_PROJECT_BASE_SECRET_KEY=myBASEkeyValue123
```

## 2. 加密

```shell
cd envtools
go build main.go
./main -d="待加密的字符串" -k="环境变量名"
```

```shell
# ./main -d="123.456" -k="XY_REDIS_AUTH"
plaintext:
	123.456
ciphertext:
	FH3Djy1UJiv2y5CrpDQzty
Linux:
	export XY_REDIS_AUTH=FH3Djy1UJiv2y5CrpDQzty
Windows:
	set XY_REDIS_AUTH=FH3Djy1UJiv2y5CrpDQzty


testGetenv: XY_REDIS_AUTH = 123.456
```

得到加密内容到环境中执行即可.

## 3. 解密

```shell
export XY_REDIS_AUTH=FH3Djy1UJiv2y5CrpDQzty
./main -k=XY_REDIS_AUTH
# testGetenv: XY_REDIS_AUTH = 123.456
```

## 4. 应用示例


```go
package main

import (
	"fmt"

	"github.com/fufuok/utils"
)

const (
	// 项目基础密钥 (环境变量名)
	BaseSecretKeyName = "FF_PROJECT_1_BASE_SECRET_KEY"

	// 用于解密基础密钥值的密钥 (编译在程序中)
	BaseSecretSalt = "123"

	// Redis Auth 短语环境变量 Key
	RedisAuthKeyName = "PROJECT_1_REDIS_AUTH"
)

type config struct {
	BaseSecret string
	RedisAuth  string
}

var Conf config

func init() {
	// 前置: 假如项目环境中已经执行了下面的配置
	// export FF_PROJECT_1_BASE_SECRET_KEY=EnUNZ1FkdnsvWXTukDe4FiwhLkw5eMmjGgAYNqYwB9zn
	// export PROJECT_1_REDIS_AUTH=FH3Djy1UJiv2y5CrpDQzty
	// 1. 项目 git 中不会出现明文信息
	// 2. 运行环境中也不会见到明文信息, 也不能通过环境变量值解密

	// 从环境变量中读取, 用程序中固化的密钥解密, 得到我们的基础密钥是: myBASEkeyValue123
	Conf.BaseSecret = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)

	// 用 BaseSecret(或基于此的密钥) 解密其他项目配置
	Conf.RedisAuth = utils.GetenvDecrypt(RedisAuthKeyName, Conf.BaseSecret)
}

func main() {
	// 业务中连接 Redis 就可以用 Conf.RedisAuth
	// Redis Auth: 123.456
	fmt.Println("Redis Auth:", Conf.RedisAuth)
}
```







*ff*