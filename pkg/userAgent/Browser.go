package userAgent

import (
	"regexp"
	"strings"
)

// DeviceType 用于判断设备类型的正则表达式
var DeviceType = map[string]*regexp.Regexp{
	"windows": regexp.MustCompile(`Windows NT (\d+\.\d+)`),
	"mac":     regexp.MustCompile(`Mac OS X (\d+_\d+(_\d+)?)`),
	"ios":     regexp.MustCompile(`CPU (iPhone )?OS (\d+_\d+(_\d+)?) like Mac OS X`),
	"iphone":  regexp.MustCompile(`iPhone`),
	"ipad":    regexp.MustCompile(`iPad`),
	"android": regexp.MustCompile(`Android (\d+(\.\d+)?)`),
	"linux":   regexp.MustCompile(`Linux`),
}

type BrowserInfo struct {
	//平台信息
	PlatformInfo platformInfo
	// 渲染引擎
	Engine string
	// 浏览器名称
	Browser string
	// 浏览器版本
	Version    string
	Extensions map[string]string
}
type platformInfo struct {

	// 设备类型，如：mobile、tablet、desktop
	DeviceType string
	Detail     map[string]string
}

//todo split 在多个空格时，需要处理

func NewBrowserInfo(ua string) *BrowserInfo {
	info := &BrowserInfo{
		Extensions: make(map[string]string),
	}
	// 筛选从(开始的字符串,到)结束的字符串
	platform := strings.Split(ua, "(")[1]
	tmp := strings.Split(platform, ")")
	platform = tmp[0]

	//解析平台信息
	info.parsecPlatform(platform)

	tmp = strings.Split(strings.TrimSpace(strings.Join(tmp[1:], "")), " ")
	//解析渲染引擎
	info.Engine = tmp[0]
	//解析浏览器信息
	info.Browser = tmp[1]
	tmp = strings.Split(info.Browser, "/")
	info.Browser = tmp[0]
	info.Version = tmp[1]

	//解析浏览器版本

	return info
}

func (b *BrowserInfo) parsecPlatform(platform string) {
	//去除括号
	platform = strings.Replace(platform, "(", "", -1)
	platform = strings.Replace(platform, ")", "", -1)
	for name, reg := range DeviceType {
		match := reg.FindStringSubmatch(platform)
		if len(match) > 0 {
			b.PlatformInfo.DeviceType = name
			break
		}
	}
	// 根据不同的设备类型，解析不同的信息
	switch b.PlatformInfo.DeviceType {
	case "windows":
		args := strings.Split(platform, ";") //Windows NT 10.0; Win64; x64; rv:88.0

		// Windows平台的处理逻辑
		if len(args) >= 3 {
			b.PlatformInfo.Detail = map[string]string{
				"OS":    "Windows",
				"OSVer": strings.TrimSpace(strings.Replace(args[0], "Windows NT", "", -1)),
				"Arch":  strings.TrimSpace(strings.TrimSpace(args[1]) + "; " + strings.TrimSpace(args[2])),
			}
		}
		//todo 更多不同平台的处理逻辑
	default:
		b.PlatformInfo.Detail = map[string]string{
			"OS": "Unknown",
		}
	}

}
