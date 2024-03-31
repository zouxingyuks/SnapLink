package ipSearcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	client = new(http.Client)
)

type Info struct {
	Country   string `json:"country"`
	ShortName string `json:"short_name"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Area      string `json:"area"`
	Isp       string `json:"isp"`
	Net       string `json:"net"`
	Ip        string `json:"ip"`
	Code      int    `json:"code"`
	Desc      string `json:"desc"`
}

func IPV4(ip string) (info *Info, err error) {
	info = new(Info)
	// https://ip.useragentinfo.com/jsonp?ip=59.164.141.201
	url := fmt.Sprintf("https://ip.useragentinfo.com/jsonp?ip=%s", ip)
	method := "GET"
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
func IPV6(ip string) (map[string]string, error) {
	panic("implement me")
}
