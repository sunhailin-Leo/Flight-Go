package main

import (
	"github.com/go-resty/resty"
	"github.com/tidwall/gjson"
	"time"
)

const (
	cityNameCodeVersion string = "267040"
	cityNameCodeURL            = "https://tce.alicdn.com/api/data.htm?ids=" + cityNameCodeVersion
	UserAgent           string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"
)

// 初始化一个城市名和城市代码的映射
var cityNameCode = make(map[string]string)

// 初始化城市名和城市代码的数据
func initCityNameCodeData() {
	dataResp, err := resty.New().R().
		SetHeader("user-agent", UserAgent).
		Get(cityNameCodeURL)
	if err != nil {
		logger.Fatalf("[Flight-Go]初始化数据接口异常, 错误原因: %v", err)
	}
	if dataResp.String() != "" {
		jsonData := gjson.Parse(dataResp.String())
		cityArray := jsonData.Get(cityNameCodeVersion).Get("value").Get("cityArray").Array()[1:]
		for _, cities := range cityArray {
			cityDD := cities.Get("tabdata").Array()
			for _, cityData := range cityDD {
				for _, city := range cityData.Get("dd").Array() {
					cityNameCode[city.Get("cityName").String()] = city.Get("cityCode").String()
				}
			}
		}
		logger.Info("[Flight-Go]初始化数据成功!")
	}
}

// 时间戳转时间
func timestampToTime(timestamp int64) string {
	if timestamp != 0 {
		return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
	} else {
		return "--:--"
	}
}
