# XY.DataRouter

## v1.103.13.22062408

- 优化调试模式
- 确保 2 个预格式化的全局时间字段值同步

## v1.103.12.22062113

- 允许配置 `es_post_batch_mb` 为小数
- 运行状态增加 `ESIndexingRate(/s)`
- 增加支持 ES8 根证书认证方式
- 增加 `/es/health` 展示 ES 集群健康状态

## v1.103.11.22060515

- 增加支持 ES8 认证配置
- 增加 ES 批量写入耗时调试日志
- 优化 ES 查询参数错误消息
- ES 搜索接口增加 `timeout_ms` 参数, 支持请求超时, 默认: `30s`
- 跳过 `gnet.v2` Windows 使用限制
- 调试模式允许设置日志为无着色

## v1.103.10.22053013

- 优化 CI: 增加守护参数
- 优化 JSON 包: 合并常用 JSON 包到 utils
- 增加从环境变量中获取转发地址 `ForwardHost`
- 修正 ES 响应失败时的结果解析

## v1.103.9.22052316

- 允许在 ES 文档更新/删除时使用 `refersh=true`

## v1.103.8.22052313

- 优化结构体命名
- 优化日志包定义

## v1.103.7.22052312

- 增加动态调整 ES 排队数量
- 优化数据处理池变量命名及运行状态字段

## v1.103.6.22051212

- 增加 ES 写入熔断器

## v1.103.5.22050717

- 升级 `gnet`, v1 到 v2
- 优化日志

## v1.103.4.22050616

- 设置 ES 搜索时接口忽略不可用的索引

## v1.103.3.22050111

- 重构 ES 客户端版本兼容代码
- 增加 ES 数据更新和删除接口 (单文档)
- 优化错误响应结果, 增加返回详细错误

## v1.103.2.22041414

- ES 查询日志增加请求时间, URI 等字段
- 替换 `xsync.NewMap` 为 `xsync.NewMapOf`
- 优化配置文件变化监控

## v1.103.1.22032818

- 增加接口配置 `es_pipeline` 参数, 指定写入 ES 时使用的 Pipeline
- 增加接口配置 `es_disable_write` 参数, 可选不写 ES, 数据仅分发给第三方接口

## v1.103.0.22032111

- 应用泛型

## v1.102.7.22031111

- 简化 `DataRouter` 结构体
- 优化数据项计数
- 优化心跳日志和日志索引

## v1.102.6.22030917

- 优化数据推送配置项

## v1.102.5.22030810

- 升级 Req 支持重试, 增加相应配置
- 升级依赖

## v1.102.4.22022515

- 规范 Init 方法命名, 重构 service 目录结构
- 修改 `/sys/status` 为 `/sys/stats`, 优化变量命名
- 全局协程池增加日志记录器
- 数据分发 PostAPI 增加单次推送最大记录数和最大字节数限制

## v1.102.3.22022410

- 修正 ES 批量写入时采样错误日志记录失败
- 配置 `req` 日志记录器为 `zerolog`
- 数据代理通道日志记录器遵循主配置文件变化
- JSON 响应默认字符集 `utf-8`
- 规范 `Web` 目录结构
- 配置变化时重置文件监听时间间隔
- 优化心跳日志

## v1.102.2.22022111

- 数据分发到多个平台时, 并发推送, 升级 `req v3` 默认使用 `HTTP/2`
- JSON 数据格式化优化
- 优化对无效 ES 数据的排除
- 接口默认启用 `HTTP/1.1`

## v1.102.1.22021616

- 规范 API 响应函数名称和配置项
- 升级 ES 客户端 SDK 为 `v7.17.0`

## v1.102.0.22011818

- 设置缓冲池最大支持 1MiB 的字节切片复用
- 重构配置文件, 配置项分组
- ES 查询语句有误时返回原因
- 接口数据拷贝时使用 bytespool

## v1.101.4.22011717

- 增加数据项初始化时更多的可选函数, 供不同场景优选
- 升级依赖

## v1.101.3.22011417

- 增加 ES 查询多级代理功能:
  - 加密头部携带客户端 IP, 中间代理自动转发
  - 自动识别客户端公网出口, 适配接口 IP 白名单验证, 与直接访问接口行为一致
- 增加 Web 配置项, 客户端 IP, 信任代理白名单等, nginx 反向代理场景适用
- 增加 ES 查询接口日志
- 增加 404 路由

## v1.101.2.22011217

- 增加 `/es/count` 接口, 获取符合条件的索引文档数
- 重构 ES 查询相关代码, 池化, 规范调用
- 重构接口路由目录结构
- 优化数据项相关方法代码, 更新基准测试
- 取消接口返回详细错误信息
- 升级依赖

## v1.101.1.22010721

- 搜索接口和数据写入兼容 ES 7.x 
- 增加配置项 `es_optional_write` 可动态调整基于排队值的繁忙比率定义
  - 繁忙时设置为 `es_optional_write` 的接口数据将不写 ES
  - 增加运行时状态 `ESDataItemDiscards` 显示繁忙时被丢弃的数据项(UDP: 1 个数据, TCP: >=1 个数据)
- 调整可选写入 ES 状态更新的时间间隔, 由 `1minute` 改为 `500ms`

## v1.101.0.22010709

- 调整配置项:
  - 无效 JSON 数据日志由 `Warn` 改为 `Info`
  - ES 慢查询默认值由 `10s` 改为 `5s`
  - ES 默认写入时间间隔由 `0.5s` 改为 `1s`
  - ES 批量写入最小排队数最小值由 `100` 改为 `10`
  - 增加 `es_retry_on_status` ES 重试状态码, 默认: `[502, 503, 504]`
  - 增加 `es_max_retries` ES 重试次数, 默认: `3`
  - 增加 `es_disable_retry` 是否禁止 ES 重试, 默认: `false`
  - 移除 ES 批量写入状态码重试配置
- 运行时状态增加:
  - JSON 库信息
  - ES 服务端和客户端版本信息
- ES `Transport` 使用 `fasthttp`
- 接口响应增加 `APISuccessBytes` 快捷方法
- ES 客户端 SDK 由 `v6.8.10` 升级为 `v7.16.0`
- ES 搜索接口结果解析性能优化
- ES 批量写入方法代码重构

## v1.100.67.21120909

- 启用 HTTPS

## v1.100.66.21120809

- 应用 `gnet` 补丁

## v1.100.65.21120707

- 升级 `gnet` 包

## v1.100.63.21113018

- 增加 UDP 获取客户端 IP 的接口: 发送 2 个字节的数据到 UDP 接口即可

## v1.100.62.21112511

- 优化配置文件版本信息

## v1.100.61.21112321

- 接口不存在的错误日志级别由 `ERROR` 调整为 `INFO`

## v1.100.60.21112211

- 优化数据代理通道:
  - 增加数据发送客户端数量, 默认为 CPU * 2, 相应减少每客户端队列大小
  - 客户端发送队列超限后, 强制重新建立连接
  - 服务端接收数据开启异步, 使用 ants 管理协程
  - 开启池化, 优化日志

## v1.100.59.21110119

- 新增 ES 繁忙时丢弃指定索引数据的策略
- 修正 ES 批量提交待处理计数

## v1.100.58.21110111

- 全面池化

## v1.100.56.21102822

- 增加通道数据压缩传输
- 使用 [github.com/fufuok/bytespool](https://github.com/fufuok/bytespool) 替换 `bytebufferpool`

## v1.100.55.21102323

- 优化对象复用

## v1.100.54.21102111

- 减少不必要的指针间接引用
- 全局时间优化

## v1.100.50.21091809

- 调整 `main` 包结构, 增加版本信息
- 优化 `Makefile` 
- 优化 `init()` 减少内部相互依赖, 明确初始化顺序

## v1.100.45.21090909

- 内嵌 `favicon.ico`
- 增加 `/client_ip` `/server_ip` 显示客户端来访 IP 和当前服务器 IP
- 升级依赖包

## v1.100.44.21082808

- 使用 `xsync.NewMap` 替代 `cmap.New`, `xsync.Counter` 替代 `atomic`
- 接口不存在日志记录客户端 IP

## v1.100.43.21082020

- 升级 `go1.17` `utils 0.2.0` 等

## v1.100.42.21080808

- 稳定版本
- 升级依赖包, 规范注释

## v1.100.41.21071515

- 增加 `TunnelDataTotal` `TunnelTodoSendCount` 等, 规范统计字段名称
- 增加配置项 `TunSendQueueSize` 发送数据最大队列长度, 默认: `8192`, 公网大丢包可能超过队列长度而丢弃数据
- `Tunnel` 数据发送不设置超时

## v1.100.40.21071323

- 数据代理由 `WsHub(Websocket)` 替换为 `Tunnel(aRPC)`

## v1.100.31.21070417

- 新增: 无限缓冲信道最大缓冲数量配置(可选) `data_chan_max_buf_cap`
  - 默认为 `0`, 无限大
  - 当前配置 `500000`, 超过限制 `DataChanSize(初始化大小, 默认 50) + DataChanMaxBufCap` 丢弃数据
  - 消费长时间慢于生产, 避免累积数据占用内存过大, 不应该出现该状态 (`WsHubQueueDiscards`)

## v1.100.30.21070323

- 新增: 多级数据代理功能, 解决海外数据上报网络问题
  1. 程序相同, 部署到各地, 可选开启代理 `-f 1.2.3.4:1234`
  2. 就近响应 HTTP/UDP 数据请求 (需配合智能 DNS, 域名解析到就近服务器)
  3. 走专线/环网将数据传递给上联服务器 `1.2.3.4` (可多级传递)
  4. 上联服务器接收到数据后, 在本地继续完成剩下的数据处理和分发逻辑

## v1.100.18.21063012

- 修正: ES 搜索接口中指针在 `go-json` 引发异常

## v1.100.17.21062911

- 增加: 数据处理待办计数等统计项
- 优化: 数据处理创建工作单元时, 排队情况下也不阻塞

## v1.100.16.21062717

- 增加可选配置: `reduce_memory_usage` 减少内存占用(可能增加CPU占用), 默认关闭

## v1.100.15.21062616

- ES 批量写入日志级别配置方式调整
- 增加接口请求黑名单功能

## v1.100.14.21062300

- 优化组路由代码, 严格路由模式, 默认回应请求后立即关闭连接
- 升级依赖到最新, 扩充时间轮精度
- 增加 HTTP 错误请求计数

## v1.100.10.21060500

- 批量提交支持更多数据形式, `/v1/apiname/bulk` 可 POST 多条 JSON 数据 

  1. 现支持数据间用 `=-:-=` 分隔, 如: `{"a":1}=-:-={"b":2}`

  2. 增加支持列表数据, 如: `[{"a":1},{"b":2}]`

  3. 增加支持混搭,如: `{"a":1}=-:-={"b":2}=-:-=[{"c":3},{"d":4}]`

  4. 任意 JSON 格式, 有无美化样式, 换行, 空格都支持, 如:

     ```json
      {
         "a" :
       1 , "b":[
           2,
           3
       ]
     } =-
     :- =[ {"c" :"d" }, {"e":5}] =-:-=
     ```

- 优化 HTTP 接口成功返回值

- 调整初始化顺序, 守护进程更干净

- 启用时间轮, 优化定时器

- 调整关闭写入 ES 时代码逻辑, 数据处理允许的并发最小值调整: `10`

## v1.100.9.21052417

- 清理一些注释和调试代码

## v1.100.8

**2021-05-23**

- 里程碑版本

## v1.100.7

**2021-05-21**

- 格式化 `sys/status` 输出
- 接口必填字段检查逻辑变更: 允许为空字符串, 只要字段存在即通过
- 适配获取远程配置接口加密新方案

## v1.100.6

**2021-05-20**

- ES 批量写入相关参数调整:
  - 并发 `es_bulker_size` 最低值允许为 `5`, 默认为 `30`, 当前设置为 `5`
  - 最大排队数 `es_bulker_waiting_limit` 最低值 `100`, 默认 `500`, 当前为默认设置, 理论最大内存占用 `530*5MB`
  - 附批量写入策略(任一条件满足即开始写):
    - `es_post_batch_num` ES 单次批量写入最大条数, 默认: `4500`, 当前为默认设置
    - `es_post_batch_mb` ES 单次批量写入最大字节数, 单位: `MB`, 最低: `1`, 默认: `5`, 当前为默认设置
    - `es_post_max_interval` ES 单次批量最大写入时间间隔, 单位: `ms`, 最低: `50`, 默认: `500`, 即 `1` 秒至少 `2` 次
- 增加 ES 批量写入失败时记录数据详情开关, `log.es_bulk_error`
- 增加配置项 `es_reentry_codes` 
  - ES 批量写入错误时, 配置中的状态码会重新进入排队, 比如: `[429]` es_rejected_execution
  - 会等待一个提交时间周期, 默认 `500` 毫秒
  - 注: 有重复数据风险, 默认未启用
- 杂项
  - 数据体 `body` 相关优化
  - JSON 添加系统字段效率优化
  - 升级 `fiber@v2.10.0` `gnet@v1.4.4` `gjson@v1.8.0` `utils@v0.1.6`

## v1.100.2

**2021-05-19**

- 增加数据处理并发限制和 ES 写入任务并发限制, 超过阈值排队, 超过最大排队数丢弃
- 新增并发数, 丢弃数等监控指标

## v1.100.1

**2021-05-18**

- 默认关闭 ES 写入重试机制, 可配置开启
- ES 无法连接时允许启动程序, 接收请求和分发并重试连接, 确保报警平台接入服务正常
- 增加 ES 写入开关, 关闭后仅保持数据分发和请求响应, 数据不进 ES
- 增加 ES 写入工作协程数和写入错误数统计
- 增加 ES 压力可控配置项, ES Bulk 协程数达到阈值时, 丢弃新数据
- 增加 HTTP 接口按来访 IP 限流, UDP 接口未限流, 配置项
- 增加 HTTP / UDP 请求简单计数, 探针包除外
- 增加访问非法接口日志, 记录来访 URI 和 IP
- 丰富运行状态统计:
  - `/sys/status` 
    - 应用启动时间, 协程数, 系统信息, 版本信息等
    - 内存占用情况
    - 接口请求数
    - ES 数据队列长度和堆积数, ES 数据条数, pending 的写入协程, 写入错误次数
  - `/sys/status/queue` 各接口数据队列长度和堆积数
  - `/debug/statsviz` 重要指标可视化

## v1.100.0

**2021-05-17**

- 新分支: 无 Redis 独立运行版本

## v1.41.12

**2021-05-13**

- 请求 `body` 直接沿用 `bytes`

## v1.41.11

**2021-05-12**

- 使用 `panjf2000/ants`, `arl/statsviz`
- 升级 `fiber@v2.9.0`
- 默认参数调整: 减少 Redis 连接池数和数据分发协程数

## v1.41.10

**2021-05-01**

- 使用 `utils/json`

## v1.41.8

**2021-04-27**

- 升级依赖包: `utils`

## v1.41.6

**2021-04-24**

- 升级依赖包: `utils` `fiber` `gjson` `go-redis`
- 更新环境加密算法
- 修正 oldAPI: `copy c.Body()`

## v1.41.4

**2021-04-16**

- `set BodyLimit=500M`
- 增加 `esSearch` 错误响应日志
- 改用 `json-iterator`

## v1.41.1

**2021-04-15**

- `copy c.Body()`

## v1.41.0

**2021-04-14**

- 规范版本号
  - `v1.33.4` 即原来 `v3.3.4` 为 gin 最后版本
  - `v1.40.0` 即原来 `v4.0.0` 为 fiber 起始版本
- 启用 `utils`

## v4.0.2

**2021-04-11**

- 启用 `go-json`

## v4.0.1

**2021-04-10**

- 修正接口名称值传递

## v4.0.0

**2021-04-09**

- `gin` 改为 `fiber`

## v3.3.4

**2021-04-06**

- 日志库由 `gxlog` 改为 `zerolog`

## v3.3.2

**2021-04-01**

- 日志策略调整, 日志格式优化
- `APIException` 入参顺序调整
- 增加 `/sys/check` 查看来源 IP 是否为白名单项
- 升级依赖包版本

## v3.3.1

**2021-02-24** 更新

- 配置重载时弃用 `IRQ chan`, 使用 `context`
- 调整一些函数和变量名

## v3.3.0

**2021-02-23** 更新

- 周期性任务时间校准

## v3.2.9

**2021-02-21** 更新

- 增加心跳日志, 用于监控报警
- `/sys/status` 带格式输出, 方便命令行查看

## v3.2.8

**2021-02-08** 更新

- 修正加密密钥长度

## v3.2.7

**2021-02-07** 更新

- 增加环境变量加密工具

## v3.2.6

**2021-02-06** 更新

- 安全的 `string` 零拷贝: `S2B`

## v3.2.5

**2021-02-05** 更新

- 增加 `S2B` `B2S` 助手函数, 零拷贝类型转换, 用于高频率转换的场景
- 远程配置无变化时, 不重写本地配置文件
- 系统状态增加当前系统配置版本及签名

## v3.2.4

**2021-02-04** 更新

- `speed_report` 删除 `data.node_line_type` 字段
- 减少 `JSON` 校验

## v3.2.3

**2021-02-03** 更新

- `speed_report` 临时方案, 单独处理接口数据, 强制转换乱码

## v3.2.2

**2021-01-31** 更新 [*]

- 自动获取远程系统配置
  - 配置在报警平台数据源在线管理
  - 新配置 `2` 分钟内生效, 配置文件监控间隔默认改为 `1` 分钟
  - 增加相应可选配置项
- 新增 `restart_main` 配置项, 可以在报警平台数据源修改配置以重启所有接口程序
- 优化锁, 分离成公共包 `redislock`
- 服务器状态分离成 `2` 个部分
  - `/sys/status` 展示关注的重要指标: 键数量, 待处理键数量, 队列数, 连接池使用率等.
    - TODO: 监控报警
  - `/sys/redis` 展示 `redis` 的 `info` 详情
- 关联性强的 `Redis` 操作调整为 `Lua` 脚本
- 调整了一些函数名称
- 升级所有依赖包到最新版

## v3.2.1

**2021-01-18** 更新

- 检测 `JSON` 兼容性优化
- 日志输出格式优化

## v3.2.0

**2021-01-15** 更新

- 调整 `Redis` `PoolSize` 默认值
- ES 搜索慢日志记录来访 IP

## v3.1.8

**2021-01-14** 更新

- 优化 `JSON` 读写性能

## v3.1.7

**2021-01-11** 更新

- 修正时间轮泄漏
- 更多可选的核心配置参数
- 增加请求头: `User-Agent: XY.DataRouter/3.1.7.21011111`

## v3.1.6

**2020-12-31** 更新

- ES 查询增加记录慢日志.

## v3.1.5

**2020-12-30** 更新

- 查询接口增加 `Scroll` 支持.
- ES 白名单增加注释.

## v3.1.3

**2020-12-26** 更新

- `INFO` 级别时记录 `esBulk` 错误
- 错误 JSON 数据日志级别调整为 `INFO`
- UDP 取索引名时直接替换冒号
- 新增配置:
    - HTTP 响应码日志记录, 默认: `500`, 即大于等于 `500` 的响应码时记录日志
    - 慢响应日志, 默认大于 `1` 秒时记录
    - `PProfAddr` 配置项

## v3.1.1

**2020-12-03** 更新

- 写入 ES 时, 固定 `_type` 为 `_doc`

## v3.1.0

**2020-11-23** 更新

- POST 接口增加请求体压缩支持
    - `/v1/apiname/gzip`
    - `/v1/apiname/bulk/gzip`
- 配置增加按月或不分割 ES 索引支持: `es_index_split`
    - `none` 不切割, 如: `api_name`
    - `day` 按天切割(默认), 如: `api_name_201123`
    - `month` 按月切割, 如: `api_name_2011`
    - `year` 按年切割, 如: `api_name_20`

## v3.0.7

**2020-11-11** 更新

- 完善 UDP / HTTP 端口, 加入配置文件

## v3.0.6

**2020-11-10** 更新

- 用端口区分是否回包, 快速完成应答环节:
  - 探针请求均应答字符: `1`
  - `39300` 端口不应答, 即服务端不会给客户端回包
  - `39400` 端口会应答, 接口收到请求数据时, 固定回包数据为字符: `1`
  - 丢弃规则:
    - 未收到数据, 网络丢包
    - 数据不是合法 JSON 键值对字符串
    - 数据中未包含 `_x` 字段, 或该字段值不符合要求
    - 接口配置中有限定必有字段, 但数据中未包含该字段

## v3.0.5

**2020-11-06** 更新

- 增加 UDP 探针(Echo)服务, 用于检测 UDP 服务是否正常
- UDP 接口应答规则:
  - 默认不应答, 即不回包
  - 可在配置文件中指定接口名需要回包(接口名为 `_x` 字段值)
  - 应答为一个字符: `0` / `1` 或无回应.

    - `1` 表示接收到的数据格式正确
    - `0` 表示接收到的数据格式有误, 缺少 `_x` 或其他必有字段(配置)
    - 无回应, 表示丢包或数据非法

## v3.0.4

**2020-11-04** 更新

- 优化 UDP 数据上报接口性能
  - 增加内存队列
  - 原子时间优化
  - 使用 `gnet` 网络包

## v3.0.0

**2020-10-21** 更新

- 新增 UDP 数据上报接口

## v2.2.0

**2020-09-25** 更新

- 正式启用包仓库(代码无变化)

## v2.1.0

**2020-09-22 更新**

- `GetGlobalTime()` 增加原子时间, 避免多服务器时间不同步引发意外
- 使用 `imroc/req` 库
- 检查必有字段逻辑优化
- `DEBUG` 时开启 `rquest` 日志

## v2.0.5

**2020-09-13 更新**

- `PostAPI` 增加锁机制, 优化锁变量命名

## v2.0.3 - 2.0.4

**2020-09-01 更新**

- 更新包名: `services` -> `service`, `utils` -> `util`
- 更新 `common.go` -> 包名同名文件
- `service_status` -> `running_status`
- `UPDATE.md` -> `CHANGELOG.md`
- 一些变量名称优化

## v2.0.1 - 2.0.2

**2020-08-25 更新**

- 配合 CI 集成调试, 无更新

## v2.0.0

**2020-08-24 更新**

- 里程碑版本

## v1.2.0

**2020-08-19 更新**

- CI
- 增加 ES 搜索中件间, IP 白名单策略
- 生产环境错误日志仅记录到文件

## v1.1.2

**2020-08-15 更新**

- Makefile
- 移动项目文件到新目录结构

## v1.1.1

**2020-08-11 更新**

- gin: + gzip; - endless

**2020-08-06 更新**

- 监听配置文件变化: 配置文件内容更新后, 自动热加载配置 (每5分钟)
- 监听程序变化: 新程序覆盖后自动重载新程序 (每5分钟)
- 自守护模式运行: 直接运行 `./xydatarouter`, 自动后台运行并守护自身
- 修正接口某些情况下返回多组结果的问题
- 日志相关格式优化(非 debug 模式下): 不记录 Web 请求日志; Warn 日志不显示文件信息

## v1.1.0

**2020-08-01 更新**

- 重构项目组织结构

## v1.0.2

**2020-07-27 更新:**

- 规范 API 返回值, 返回值增加 `id`, 与 `ok` 意义相同
- 配置文件增加 `with_sys_field` 参数, 用于推送数据到第三方时是否携带系统内置字段  
    - `_ctime` 接口收到数据请求的时间, 0 时区的北京时间值
    - `_gtime` 与 `_ctime` 相同, 正常时区时间
    - `_cip` 接口获取到的客户端IP
