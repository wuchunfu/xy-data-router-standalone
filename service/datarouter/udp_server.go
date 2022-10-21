package datarouter

import (
	"log"
	"net"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/sync/errgroup"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

var (
	// UDP 返回值
	outBytes = []byte("1")
)

func initUDPServer() {
	fn := udpServer
	if conf.Config.UDPConf.Proto == "gnet" {
		fn = udpServerG
	}

	common.Log.Info().
		Str("raddr", conf.Config.UDPConf.ServerRAddr).
		Str("rwaddr", conf.Config.UDPConf.ServerRWAddr).
		Str("proto", conf.Config.UDPConf.Proto).
		Msg("Listening and serving UDP")

	eg := errgroup.Group{}
	eg.Go(func() error {
		return fn(conf.Config.UDPConf.ServerRAddr, false)
	})
	eg.Go(func() error {
		return fn(conf.Config.UDPConf.ServerRWAddr, true)
	})
	if err := eg.Wait(); err != nil {
		log.Fatalln("Failed to start UDP Server:", err, "\nbye.")
	}
}

// 标准的 UDP 服务
func udpServer(addr string, withSendTo bool) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	// 收发缓冲区
	// _ = conn.SetReadBuffer(1024 * 1024 * 20)
	// _ = conn.SetWriteBuffer(1024 * 1024 * 20)

	// UDP 接口并发读取数据协程
	for i := 0; i < conf.Config.UDPConf.GoReadNum; i++ {
		go udpReader(conn, withSendTo)
	}

	select {}
}

// UDP 数据读取
func udpReader(conn *net.UDPConn, withSendTo bool) {
	buf := make([]byte, conf.UDPMaxRW)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil || n == 0 {
			return
		}

		clientIP := clientAddr.IP.String()

		if withSendTo || n < jsonMinLen {
			out := outBytes
			if n == 2 {
				// 返回客户端 IP
				out = utils.S2B(clientIP)
			}
			_ = common.GoPool.Submit(func() {
				writeToUDP(conn, clientAddr, out)
			})
		}

		if n >= jsonMinLen {
			item := schema.NewSafeBody("", clientIP, buf[:n])
			_ = common.GoPool.Submit(func() {
				if !saveUDPData(item) {
					item.Release()
				}
			})
		}
	}
}

// UDP 应答
func writeToUDP(conn *net.UDPConn, clientAddr *net.UDPAddr, out []byte) {
	_, _ = conn.WriteToUDP(out, clientAddr)
}

// 校验并保存数据
func saveUDPData(item *schema.DataItem) bool {
	// 请求计数
	UDPRequestCount.Inc()

	if len(conf.ESBlackListConfig) > 0 && utils.InIPNetString(item.IP, conf.ESBlackListConfig) {
		common.LogSampled.Info().Str("method", "UDP").Msg("非法访问: " + item.IP)
		return false
	}

	// 接口名称与索引名称相同, 存放在 _x 字段
	esIndex := getUDPESIndex(item.Body, conf.UDPESIndexField)
	if esIndex == "" {
		return false
	}

	// 接口配置检查
	apiConf, ok := conf.APIConfig[esIndex]
	if !ok || apiConf.APIName == "" {
		common.LogSampled.Info().
			Str("client_ip", item.IP).Str("udp_x", esIndex).Int("len", len(esIndex)).
			Msg("api not found")
		return false
	}

	// 必有字段校验
	if !common.CheckRequiredField(item.Body, apiConf.RequiredField) {
		return false
	}

	// 保存数据
	item.APIName = esIndex
	schema.PushDataToChanx(item)

	return true
}

// 获取 ES 索引名称
func getUDPESIndex(body []byte, key string) string {
	index := gjson.GetBytes(body, key).String()
	if index != "" {
		return utils.ToLower(utils.Trim(index, ' '))
	}
	return ""
}
