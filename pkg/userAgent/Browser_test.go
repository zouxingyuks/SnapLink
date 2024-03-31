package userAgent

import (
	"fmt"
	"strings"
	"testing"
)

func equalMaps(map1, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value1 := range map1 {
		value2, ok := map2[key]
		if !ok || value1 != value2 {
			return false
		}
	}

	return true
}

func cmp(got, want *BrowserInfo) bool {

	if strings.Compare(got.Browser, want.Browser) != 0 {
		fmt.Println("got.Browser:", got.Browser, "want.Browser:", want.Browser)
		return false
	}
	if strings.Compare(got.Engine, want.Engine) != 0 {
		fmt.Println("got.Engine:", got.Engine, "want.Engine:", want.Engine)
		return false
	}
	if strings.Compare(got.Version, want.Version) != 0 {
		fmt.Println("got.Version:", got.Version, "want.Version:", want.Version)
		return false

	}

	if strings.Compare(got.PlatformInfo.DeviceType, want.PlatformInfo.DeviceType) != 0 {
		fmt.Println("got.PlatformInfo.DeviceType:", got.PlatformInfo.DeviceType, "want.PlatformInfo.DeviceType:", want.PlatformInfo.DeviceType)
		return false

	}
	if !equalMaps(got.PlatformInfo.Detail, want.PlatformInfo.Detail) {
		fmt.Println("got.PlatformInfo.Detail:", got.PlatformInfo.Detail, "want.PlatformInfo.Detail:", want.PlatformInfo.Detail)
		return false
	}
	if !equalMaps(got.Extensions, want.Extensions) {
		fmt.Println("got.Extensions:", got.Extensions, "want.Extensions:", want.Extensions)
		return false

	}
	return true

}

func TestNewBrowserInfo(t *testing.T) {
	type args struct {
		ua string
	}
	tests := []struct {
		name string
		args args
		want *BrowserInfo
	}{
		//firefox
		{
			name: "windows firefox",
			args: args{
				ua: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			},
			want: &BrowserInfo{
				Engine:  "Gecko/20100101",
				Browser: "Firefox",
				Version: "89.0",
				PlatformInfo: platformInfo{
					DeviceType: "windows",
					Detail: map[string]string{
						"OS":    "Windows",
						"OSVer": "10.0",
						"Arch":  "Win64; x64",
					},
				},
			},
		},
		{
			name: "linux firefox",
			args: args{
				ua: "Mozilla/5.0 (X11; U; Linux x86_64; zh-CN; rv:1.9.2.10) Gecko/20100922 Ubuntu/10.10 (maverick) Firefox/3.6.10",
			},
			want: &BrowserInfo{
				Engine:  "Gecko/20100922",
				Browser: "Firefox",
				Version: "3.6.10",
				PlatformInfo: platformInfo{
					DeviceType: "linux",
					Detail: map[string]string{
						"OS":    "Linux",
						"Arch":  "x86_64",
						"Lang":  "zh-CN",
						"OSVer": "maverick",
					},
				},
			},
		},
		{
			name: "windows chrome",
			args: args{
				ua: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		},

		{
			name: "windows edge",
			args: args{
				ua: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			},
		},
		{
			name: "mac chrome",
			args: args{
				ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		},
		{
			name: "mac safari",
			args: args{
				ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15",
			},
		},
		{
			name: "iphone safari",
			args: args{
				ua: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Mobile/15E148 Safari/604.1",
			},
		},
		{
			name: "ipad safari",
			args: args{
				ua: "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Mobile/15E148 Safari/604.1",
			},
		},
		{
			name: "android chrome",
			args: args{
				ua: "Mozilla/5.0 (Linux; Android 11; Pixel 4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBrowserInfo(tt.args.ua); !cmp(got, tt.want) {
				t.Errorf("NewBrowserInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
