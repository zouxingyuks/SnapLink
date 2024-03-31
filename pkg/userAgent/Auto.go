package userAgent

import "github.com/ua-parser/uap-go/uaparser"

var parser = uaparser.NewFromSaved()

type Info struct {
	Engine  string
	Browser string
	Device  string
	Version string
	OS      string
}

// AutoParse 自动解析,会根据ua自动解析是浏览器还是机器人
func AutoParse(ua string) (info *Info) {
	//browserInfo := &BrowserInfo{
	//	Engine:  "Gecko/20100922",
	//	Browser: "Firefox",
	//	Version: "3.6.10",
	//	PlatformInfo: platformInfo{
	//		DeviceType: "linux",
	//		Detail: map[string]string{
	//			"OS":    "Linux",
	//			"Arch":  "x86_64",
	//			"Lang":  "zh-CN",
	//			"OSVer": "maverick",
	//		},
	//	},
	//}
	client := parser.Parse(ua)
	info = new(Info)
	info.Browser = client.UserAgent.Family // 获取浏览器信息
	info.Version = client.UserAgent.Major  // 获取浏览器版本
	info.Engine = client.UserAgent.Minor   // 获取浏览器引擎信息
	info.Device = client.Device.Family     // 获取设备信息
	info.OS = client.Os.Family             // 获取操作系统信息
	return info

}
