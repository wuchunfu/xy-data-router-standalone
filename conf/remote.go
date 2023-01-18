package conf

import (
	"fmt"
	"os"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xjson/gjson"
	"github.com/imroc/req/v3"
)

// GetMonitorSource 获取监控平台源数据配置
func (c *FilesConf) GetMonitorSource() error {
	// Token: md5(timestamp + auth_key)
	timestamp := utils.MustString(time.Now().Unix())
	token := utils.MD5Hex(timestamp + c.SecretValue)

	// 请求数据源
	resp, err := req.Get(c.API + token + "&time=" + timestamp)
	if err != nil {
		return err
	}

	res := utils.B2S(resp.Bytes())
	if gjson.Get(res, "ok").Bool() {
		// 获取所有配置项数据 ["item1 ip txt", "item2 ip txt"]
		body := ""
		data := gjson.Get(res, "data.#.ip_info").Array()
		if len(data) > 0 {
			for _, x := range data {
				body += x.String() + "\n"
			}

			// 当前版本信息
			ver := GetFileVer(c.Path)
			md5New := utils.MD5Hex(body)
			if md5New != ver.MD5 {
				// 保存到配置文件
				if err := os.WriteFile(c.Path, []byte(body), 0644); err != nil {
					return err
				}
				ver.MD5 = md5New
				ver.LastUpdate = time.Now()
			}

			return nil
		}
	}

	return fmt.Errorf("数据源获取失败: %s", gjson.Get(res, "msg").String())
}
