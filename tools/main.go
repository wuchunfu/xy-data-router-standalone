// 环境变量加密工具
// go run main.go -d=Fufu
// go run main.go -d="Fufu  777"
// go run main.go -d=Fufu -k=TestEnv
// go run main.go -k=TestEnv
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/fufuok/utils/xcrypto"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	// 基础密钥 (加密) script 脚本里一致 (ff.Demo.Secret):
	// export DR_BASE_SECRET_KEY=CXPJXUoN2uAo9p5DQhnwgt
	baseSecret = "CXPJXUoN2uAo9p5DQhnwgt"
	// 项目基础密钥
	baseSecretValue = ""

	// 环境变量名(可选)
	key string
	// 待加解密内容
	value string

	// 编码用户名密码字符串
	user, password string
)

func init() {
	// 基础密钥
	baseSecretValue = xcrypto.Decrypt(baseSecret, conf.BaseSecretSalt)
	if baseSecretValue == "" {
		log.Fatalln("基础密钥解密失败, 请检查程序")
	}
}

func main() {
	flag.StringVar(&user, "u", "", "用户名")
	flag.StringVar(&password, "p", "", "密码")
	flag.StringVar(&key, "k", "envname", "环境变量名")
	flag.StringVar(&value, "d", "", "待加密字符串")
	flag.Parse()

	if user != "" || password != "" {
		fmt.Printf("url.UserPassword:\n%s\n%s\n%s\n", user, password, url.UserPassword(user, password))
	}

	if value != "" {
		// 加密
		result, err := xcrypto.SetenvEncrypt(key, value, baseSecretValue)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("\nplaintext:\n\t%s\nciphertext:\n\t%s\nLinux:\n\texport %s=%s\nWindows:\n\tset %s=%s\n\n",
			value, result, key, result, key, result)
	}

	// 解密
	result := xcrypto.GetenvDecrypt(key, baseSecretValue)
	fmt.Printf("\ntestGetenv: %s = %s\n\n", key, result)
}
