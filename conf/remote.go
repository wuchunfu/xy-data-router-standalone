package conf

import (
	"fmt"
	"os"
	"time"

	"github.com/fufuok/utils"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

// GetMonitorSource 获取监控平台源数据配置
func (c *TFilesConf) GetMonitorSource() error {
	// Token: md5(timestamp + auth_key)
	timestamp := utils.MustString(time.Now().Unix())
	token := utils.MD5Hex(timestamp + c.SecretValue)

	// 请求数据源
	resp, err := req.Get(c.API+token+"&time="+timestamp, ReqUserAgent)
	if err != nil {
		return err
	}

	res := resp.String()
	if gjson.Get(res, "ok").Bool() {
		// 获取所有配置项数据 ["item1 ip txt", "item2 ip txt"]
		body := ""
		data := gjson.Get(res, "data.#.ip_info").Array()
		if len(data) > 0 {
			for _, x := range data {
				body += x.String() + "\n"
			}

			// 当前版本信息
			ver := GetFilesVer(c.Path)
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
