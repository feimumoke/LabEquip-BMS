package hostutil

import (
	"net"
	"os"
	"runtime"
	"time"
)

var HostInfoConstant = GetHostInfo()

type HostInfo struct {
	HostName        string
	HostIp          string
	StartTime       int64
	FirstReportTime int64 //本机启动之后首次上报心跳时间
}

func GetHostInfo() *HostInfo {
	hostname, _ := os.Hostname()
	ip, _ := getLocalIPv4Address()
	return &HostInfo{
		HostName:  hostname,
		HostIp:    ip,
		StartTime: time.Now().Unix(),
	}
}

func getLocalIPv4Address() (ipv4Address string, err error) {
	//获取所有网卡
	addrs, err := net.InterfaceAddrs()
	//遍历
	for _, addr := range addrs {
		//取网络地址的网卡的信息
		ipNet, isIpNet := addr.(*net.IPNet)
		//是网卡并且不是本地环回网卡
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			//能正常转成ipv4
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return
}

func GetProcessMemoryMB() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapAlloc) / 1024 / 1024
}
