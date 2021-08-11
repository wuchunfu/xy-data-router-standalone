package service

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/myip"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

type tDataRouterStats struct {
	DataRouterQueue tChanLen
	APIQueue        tChanLen
}

type tChanLen struct {
	AllLen   int
	BufLen   int
	Discards uint64
}

var (
	// 系统启动时间
	start = time.Now()

	// InternalIPv4 服务器 IP
	InternalIPv4 string
	ExternalIPv4 string
)

func initRuntime() {
	go func() {
		InternalIPv4 = myip.InternalIPv4()
	}()
	go func() {
		ExternalIPv4 = myip.ExternalIPAny(10)
	}()
}

// RunningStatus 运行状态
func RunningStatus() map[string]interface{} {
	return map[string]interface{}{
		"DATA": dataStats(),
		"SYS":  sysStatus(),
		"MEM":  memStats(),
	}
}

// RunningQueueStatus 队列状态
func RunningQueueStatus() map[string]interface{} {
	return chanStats()
}

// 系统信息
func sysStatus() map[string]interface{} {
	return map[string]interface{}{
		"APPName":      conf.WebAPPName,
		"Version":      conf.CurrentVersion,
		"Update":       conf.LastChange,
		"Uptime":       time.Since(start).String(),
		"Debug":        conf.Config.SYSConf.Debug,
		"LogLevel":     zerolog.Level(conf.Config.SYSConf.Log.Level).String(),
		"ConfigVer":    conf.Config.SYSConf.MainConfig.ConfigVer,
		"ConfigMD5":    conf.Config.SYSConf.MainConfig.ConfigMD5,
		"GoVersion":    runtime.Version(),
		"NumCpus":      runtime.NumCPU(),
		"NumGoroutine": runtime.NumGoroutine(),
		"OS":           runtime.GOOS,
		"NumCgoCall":   runtime.NumCgoCall(),
		"InternalIPv4": InternalIPv4,
		"ExternalIPv4": ExternalIPv4,

		// HTTP 服务是否开启了减少内存占用选项
		"ReduceMemoryUsage": conf.Config.SYSConf.ReduceMemoryUsage,
		// HTTP 服务是否开启了 Keepalive
		"EnableKeepalive": conf.Config.SYSConf.EnableKeepalive,
		// Tun 数据转发地址, 为空时本地处理数据
		"ForwardTunnel": conf.ForwardTunnel,
	}
}

// 内存信息
func memStats() map[string]interface{} {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return map[string]interface{}{
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
func dataStats() map[string]interface{} {
	tunSendErrors := atomic.LoadUint64(&TunSendErrors)
	tunSendCount := atomic.LoadUint64(&TunSendCount)
	tunTotal := atomic.LoadUint64(&TunDataTotal)

	return map[string]interface{}{
		// 数据传输到 ES 处理通道繁忙状态
		"ESDataQueueAll":             esChan.Len(),
		"ESDataQueueBuf":             esChan.BufLen(),
		"ESDataQueueDiscards_______": esChan.Discards(),

		// 数据传输通道繁忙状态
		"TunnelQueueAll":             TunChan.Len(),
		"TunnelQueueBuf":             TunChan.BufLen(),
		"TunnelQueueDiscards_______": TunChan.Discards(),

		"CounterStartTime": counterStartTime,

		// 公共协程池, 不阻塞
		"CommonPoolFree":    common.Pool.Free(),
		"CommonPoolRunning": common.Pool.Running(),

		// 数据处理协程池, 排队, 待处理数据量, 丢弃数据量, 繁忙状态
		"DataProcessorTodoCount____": atomic.LoadInt64(&dataProcessorTodoCount),
		"DataProcessorDiscards_____": atomic.LoadUint64(&dataProcessorDiscards),
		"DataProcessorWorkerRunning": dataProcessorPool.Running(),
		"DataProcessorWorkerFree___": dataProcessorPool.Free(),

		// ES 总数据量, 排队, 待批量写入任务数, 丢弃任务数, 写入错误任务数, 繁忙状态
		"ESDataTotal":                utils.Commau(atomic.LoadUint64(&esDataTotal)),
		"ESBulkTodoCount___________": atomic.LoadInt64(&esBulkTodoCount),
		"ESBulkCount":                utils.Commau(atomic.LoadUint64(&esBulkCount)),
		"ESBulkDiscards____________": atomic.LoadUint64(&esBulkDiscards),
		"ESBulkErrors______________": atomic.LoadUint64(&esBulkErrors),
		"ESBulkWorkerRunning":        esBulkPool.Running(),
		"ESBulkWorkerFree__________": esBulkPool.Free(),

		// HTTP 请求数, 非法/错误请求数, UDP 请求数, Tunnel 收发数据数
		"HTTPRequestCount":           utils.Commau(atomic.LoadUint64(&HTTPRequestCount)),
		"HTTPBadRequestCount":        utils.Commau(atomic.LoadUint64(&HTTPBadRequestCount)),
		"UDPRequestCount":            utils.Commau(atomic.LoadUint64(&UDPRequestCount)),
		"TunnelRecvCount":            utils.Commau(atomic.LoadUint64(&TunRecvCount)),
		"TunnelRecvBadCount":         utils.Commau(atomic.LoadUint64(&TunRecvBadCount)),
		"TunnelDataTotal":            utils.Commau(tunTotal),
		"TunnelSendCount":            utils.Commau(tunSendCount),
		"TunnelSendErrors__________": tunSendErrors,
		"TunnelTodoSendCount_______": tunTotal - tunSendCount - tunSendErrors,
	}
}

// 数据队列信息
func chanStats() map[string]interface{} {
	stats := map[string]interface{}{}
	for item := range dataRouters.IterBuffered() {
		dr := item.Val.(*tDataRouter)
		stats[item.Key] = tDataRouterStats{
			DataRouterQueue: tChanLen{
				AllLen:   dr.drChan.Len(),
				BufLen:   dr.drChan.BufLen(),
				Discards: dr.drChan.Discards(),
			},
			APIQueue: tChanLen{
				AllLen:   dr.drOut.apiChan.Len(),
				BufLen:   dr.drOut.apiChan.BufLen(),
				Discards: dr.drOut.apiChan.Discards(),
			},
		}
	}

	return stats
}
