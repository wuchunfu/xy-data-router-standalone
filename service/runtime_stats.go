package service

import (
	"runtime"
	"time"

	"github.com/fufuok/utils"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/service/datarouter"
	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/service/schema"
	"github.com/fufuok/xy-data-router/service/tunnel"
)

var (
	// ES 索引数据临时计数和值对应的时间点
	esIndexTotal     int64
	esIndexTotalTime time.Time
)

// RuntimeStats 运行状态统计
func RuntimeStats() map[string]any {
	return map[string]any{
		"DATA": dataStats(),
		"SYS":  sysStats(),
		"MEM":  memStats(),
	}
}

// 系统信息
func sysStats() map[string]any {
	ver := conf.GetFileVer(conf.ConfigFile)
	return map[string]any{
		"APPName":      conf.APPName,
		"Version":      conf.Version,
		"GitCommit":    conf.GitCommit,
		"Uptime":       time.Since(common.Start).String(),
		"StartTime":    common.Start,
		"Debug":        conf.Debug,
		"LogLevel":     zerolog.Level(conf.Config.LogConf.Level).String(),
		"ConfigVer":    ver.LastUpdate,
		"ConfigMD5":    ver.MD5,
		"GoVersion":    conf.GoVersion,
		"NumCpus":      runtime.NumCPU(),
		"NumGoroutine": runtime.NumGoroutine(),
		"NumCgoCall":   utils.Comma(runtime.NumCgoCall()),
		"InternalIPv4": common.InternalIPv4,
		"ExternalIPv4": common.ExternalIPv4,

		// HTTP 服务是否开启了减少内存占用选项
		"ReduceMemoryUsage": conf.Config.WebConf.ReduceMemoryUsage,
		// HTTP 服务是否关闭了 keep-alive
		"DisableKeepalive": conf.Config.WebConf.DisableKeepalive,
		// Tun 数据转发地址, 为空时本地处理数据
		"ForwardHost": conf.ForwardHost,
		// 配置中是否关闭了 ES 写入
		"ConfESDisableWrite": conf.Config.DataConf.ESDisableWrite,
		// UDP 协议原型
		"UDPProto": conf.Config.UDPConf.Proto,
		// 是否启用了 HTTPS
		"HTTPS": conf.Config.WebConf.ServerHttpsAddr != "",
		// JSON 库信息
		"JSON": json.Name,

		// ES 版本信息
		"ESServerVer": es.ServerVer,
		"ESClientVer": es.ClientVer,
	}
}

// 内存信息
func memStats() map[string]any {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return map[string]any{
		// 程序启动后累计申请的字节数
		"TotalAlloc":  ms.TotalAlloc,
		"TotalAlloc_": utils.HumanBytes(ms.TotalAlloc),
		// 虚拟占用, 总共向系统申请的字节数
		"HeapSys":  ms.HeapSys,
		"HeapSys_": utils.HumanBytes(ms.HeapSys),
		// 使用中或未使用, 但未被 GC 释放的对象的字节数
		"HeapAlloc":  ms.HeapAlloc,
		"HeapAlloc_": utils.HumanBytes(ms.HeapAlloc),
		// 使用中的对象的字节数
		"HeapInuse":  ms.HeapInuse,
		"HeapInuse_": utils.HumanBytes(ms.HeapInuse),
		// 已释放的内存, 还没被堆再次申请的内存
		"HeapReleased":  ms.HeapReleased,
		"HeapReleased_": utils.HumanBytes(ms.HeapReleased),
		// 没被使用的内存, 包含了 HeapReleased, 可被再次申请和使用
		"HeapIdle":  ms.HeapIdle,
		"HeapIdle_": utils.HumanBytes(ms.HeapIdle),
		// 分配的对象数
		"HeapObjects":  ms.HeapObjects,
		"HeapObjects_": utils.Commau(ms.HeapObjects),
		// 下次 GC 的阈值, 当 HeapAlloc 达到该值触发 GC
		"NextGC":  ms.NextGC,
		"NextGC_": utils.HumanBytes(ms.NextGC),
		// 上次 GC 时间
		"LastGC": time.Unix(0, int64(ms.LastGC)).Format(time.RFC3339Nano),
		// GC 次数
		"NumGC": utils.Commau(ms.NextGC),
		// 被强制 GC 的次数
		"NumForcedGC": ms.NumForcedGC,
	}
}

// 数据处理信息
func dataStats() map[string]any {
	tunSendErrors := tunnel.SendErrors.Value()
	tunSendCount := tunnel.SendCount.Value()
	tunTotal := tunnel.ItemTotal.Value()

	// 计算 ES 索引数据速率
	esDataTotal := datarouter.ESDataTotal.Value()
	esRate := calcESIndexingRate(esDataTotal)

	return map[string]any{
		// 数据传输到 ES 处理通道繁忙状态
		"ESDataQueueAll":              datarouter.ESChan.Len(),
		"ESDataQueueBuf":              datarouter.ESChan.BufLen(),
		"ESDataQueueDiscards________": datarouter.ESChan.Discards(),

		// 数据传输通道繁忙状态
		"TunnelQueueAll":              schema.ItemTunChan.Len(),
		"TunnelQueueBuf":              schema.ItemTunChan.BufLen(),
		"TunnelQueueDiscards________": schema.ItemTunChan.Discards(),
		"DataRouterQueueAll":          schema.ItemDrChan.Len(),
		"DataRouterQueueBuf":          schema.ItemDrChan.BufLen(),
		"DataRouterQueueDiscards____": schema.ItemDrChan.Discards(),

		// 数据项统计
		"AllItemTotal":        utils.Comma(schema.ItemTotal.Value()),
		"DataRouterItemTotal": utils.Comma(datarouter.ItemTotal.Value()),
		"TunnelItemTotal":     utils.Comma(tunTotal),

		// 公共协程池, 不阻塞
		"CommonGoPoolFree":    common.GoPool.Free(),
		"CommonGoPoolRunning": common.GoPool.Running(),

		// 数据处理协程池, 排队, 待处理数据量, 丢弃数据量, 繁忙状态
		"DataProcessorDiscards______": datarouter.DataProcessorDiscards.Value(),
		"DataProcessorRunning":        datarouter.DataProcessorPool.Running(),
		"DataProcessorWaiting":        datarouter.DataProcessorPool.Waiting(),
		"DataProcessorWaitingLimit":   conf.Config.DataConf.ProcessorWaitingLimit,
		"DataProcessorFree__________": datarouter.DataProcessorPool.Free(),

		// 设置为可选写入 ES 的接口丢弃数据项计数
		"ESDataItemDiscards_________": datarouter.ESDataItemDiscards.Value(),

		// ES 总数据量, 排队, 待批量写入任务数, 丢弃任务数, 写入错误任务数, 繁忙状态
		"ESDataTotal":                 utils.Comma(esDataTotal),
		"ESBulkCount":                 utils.Comma(es.BulkCount.Value()),
		"ESBulkErrors_______________": es.BulkErrors.Value(),
		"ESBulkDiscards_____________": datarouter.ESBulkDiscards.Value(),
		"ESBulkerRunning":             datarouter.ESBulkPool.Running(),
		"ESBulkerWaiting____________": datarouter.ESBulkPool.Waiting(),
		"ESBulkerWaitingLimit":        conf.Config.DataConf.ESBulkerWaitingLimit,
		"ESBulkerFree_______________": datarouter.ESBulkPool.Free(),
		// ES 写入熔断次数
		"ESBreakerCount_____________": datarouter.ESBreakerCount.Value(),
		// 是否关闭了 ES 写入
		"ESDisableWrite_____________": datarouter.ESDisableWrite.Load(),
		// 繁忙时自动开启, 开启时所有设置了该标识的接口数据将不会写入 ES
		"ESOptionalWrite____________": datarouter.ESOptionalWrite.Load(),
		// ES 索引数据速率
		"ESIndexingRate(/s)": esRate,

		// HTTP 请求数, 非法/错误请求数, UDP 请求数, Tunnel 收发数据数
		"HTTPRequestCount":            utils.Comma(common.HTTPRequestCount.Value()),
		"HTTPBadRequestCount":         utils.Comma(common.HTTPBadRequestCount.Value()),
		"UDPRequestCount":             utils.Comma(datarouter.UDPRequestCount.Value()),
		"TunnelRecvCount":             utils.Comma(tunnel.RecvCount.Value()),
		"TunnelRecvBadCount_________": tunnel.RecvBadCount.Value(),
		"TunnelCompressTotal":         utils.Comma(tunnel.CompressTotal.Value()),
		"TunnelSendCount":             utils.Comma(tunSendCount),
		"TunnelSendErrors___________": tunSendErrors,
		"TunnelTodoSendCount________": tunTotal - tunSendCount - tunSendErrors,
	}
}

func calcESIndexingRate(n int64) (rate float64) {
	now := time.Now()
	if esIndexTotal > 0 {
		rate = float64(n-esIndexTotal) / now.Sub(esIndexTotalTime).Seconds()
		rate = utils.Round(rate, 2)
	}
	esIndexTotal = n
	esIndexTotalTime = now
	return
}
