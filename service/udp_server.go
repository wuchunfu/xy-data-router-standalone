package service

import (
	"log"
	"net"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	// UDP 返回值
	outBytes = []byte("1")
)

func initUDPServer() {
	exitUDPChan := make(chan error)

	switch conf.Config.SYSConf.UDPProto {
	case "gnet":
		go func() {
			if err := udpServerG(conf.Config.SYSConf.UDPServerRWAddr, true); err != nil {
				exitUDPChan <- err
			}
		}()
		go func() {
			if err := udpServerG(conf.Config.SYSConf.UDPServerRAddr, false); err != nil {
				exitUDPChan <- err
			}
		}()
	default:
		go func() {
			if err := udpServer(conf.Config.SYSConf.UDPServerRWAddr, true); err != nil {
				exitUDPChan <- err
			}
		}()
		go func() {
			if err := udpServer(conf.Config.SYSConf.UDPServerRAddr, false); err != nil {
				exitUDPChan <- err
			}
		}()
	}

	common.Log.Info().
		Str("raddr", conf.Config.SYSConf.UDPServerRAddr).Str("rwaddr", conf.Config.SYSConf.UDPServerRWAddr).
		Msg("Listening and serving UDP")

	err := <-exitUDPChan
	log.Fatalln("Failed to start UDP Server:", err, "\nbye.")
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
	for i := 0; i < conf.Config.SYSConf.UDPGoReadNum; i++ {
		go udpReader(conn, withSendTo)
	}

	select {}
}

// UDP 数据读取
func udpReader(conn *net.UDPConn, withSendTo bool) {
	readerBuf := make([]byte, conf.UDPMaxRW)
	for {
		n, clientAddr, err := conn.ReadFromUDP(readerBuf)
		if err == nil && n > 0 {
			if withSendTo || n < 7 {
				_ = common.Pool.Submit(func() {
					writeToUDP(conn, clientAddr)
				})
			}
			if n >= 7 {
				body := utils.CopyBytes(readerBuf[:n])
				clientIP := clientAddr.IP.String()
				_ = common.Pool.Submit(func() {
					saveUDPData(body, clientIP)
				})
			}
		}
	}
}

// UDP 应答
func writeToUDP(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	_, _ = conn.WriteToUDP(outBytes, clientAddr)
}

// 校验并保存数据
func saveUDPData(body []byte, clientIP string) bool {
	// 请求计数
	udpRequestCount.Inc()

	if len(conf.ESBlackListConfig) > 0 && utils.InIPNetString(clientIP, conf.ESBlackListConfig) {
		common.LogSampled.Info().Str("method", "UDP").Msg("非法访问: " + clientIP)
		return false
	}

	// 接口名称与索引名称相同, 存放在 _x 字段
	esIndex := getUDPESIndex(body, conf.UDPESIndexField)
	if esIndex == "" {
		return false
	}

	// 接口配置检查
	apiConf, ok := conf.APIConfig[esIndex]
	if !ok || apiConf.APIName == "" {
		common.LogSampled.Error().
			Str("client_ip", clientIP).Str("udp_x", esIndex).Int("len", len(esIndex)).
			Msg("api not found")
		return false
	}

	// 必有字段校验
	if !common.CheckRequiredField(body, apiConf.RequiredField) {
		return false
	}

	// 保存数据
	PushDataToChanx(esIndex, clientIP, body)

	return true
}
