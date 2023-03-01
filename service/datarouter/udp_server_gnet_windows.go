//go:build !linux

package datarouter

// windows 下无法使用 gnet.v2
var udpServerGnet = udpServer
