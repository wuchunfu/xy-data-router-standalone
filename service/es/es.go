package es

import (
	"fmt"
	"log"
	"strings"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/fufuok/utils/xsync"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
)

var (
	// ServerVer 版本信息
	ServerVer string

	// ServerMainVer 大版本号: 6 / 7 / 8
	ServerMainVer int

	// ServerLessThan7 大版本号小于 7
	ServerLessThan7 bool

	// BulkCount ES Bulk 批量写入完成计数
	BulkCount xsync.Counter

	// BulkErrors ES Bulk 写入错误次数
	BulkErrors xsync.Counter
)

// 首次初始化 ES 连接
func initES() {
	if err := loadES(); err != nil {
		log.Fatalln("Failed to initialize ES:", err, "\nbye.")
	}
}

// 重新初始化 ES 连接, 成功则更新连接
func loadES() error {
	// 数据转发时不涉及 ES
	if common.ForwardTunnel != "" {
		return nil
	}

	client, cfgErr, esErr := newES()
	if cfgErr != nil || esErr != nil {
		return fmt.Errorf("cfgErr: %v, esErr: %v", cfgErr, esErr)
	}

	Client = client.client

	return nil
}

func parseVersion(client esClient) error {
	resp := GetResponse()
	defer PutResponse(resp)
	resp.Response, resp.Err = client.client.Info()
	if resp.Err != nil {
		return resp.Err
	}
	if resp.Response.IsError() {
		err := fmt.Errorf("ES info error, status: %s", resp.Response.Status())
		common.Log.Error().Err(err).Msg("es.Info")
		return err
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	n, _ := buf.ReadFrom(resp.Response.Body)
	if n == 0 {
		err := fmt.Errorf("ES info error: nil")
		common.Log.Error().Err(err).Msg("es.Info")
		return err
	}

	ServerVer = gjson.GetBytes(buf.Bytes(), "version.number").String()
	ServerMainVer = utils.MustInt(strings.SplitN(ServerVer, ".", 2)[0])
	ServerLessThan7 = ServerMainVer < 7
	common.Log.Info().Str("server_version", ServerVer).Str("client_version", ClientVer).Msg("ES info")
	return nil
}
