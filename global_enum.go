package main

import "fmt"

// 公共常量
const (
	ContentTypeJson string = "application/json"
	ContentTypeForm string = "application/x-www-form-urlencoded"

	HasMeal    string = "有餐食"
	HasNotMeal string = "无餐食"

	SuperEconomyClassName string = "超级经济舱"
	EconomyClassName      string = "经济舱"
	BusinessClassName     string = "商务舱"
	FirstClassName        string = "头等舱"
)

var CabinClassMap = map[string]string{
	"Y":    EconomyClassName,
	"C":    BusinessClassName,
	"F":    FirstClassName,
	"S":    SuperEconomyClassName,
	"@S-Y": fmt.Sprintf("%s-%s", SuperEconomyClassName, EconomyClassName),
	"@S-C": fmt.Sprintf("%s-%s", SuperEconomyClassName, BusinessClassName),
}

// 国内航线查询的相关常量
const (
	PlaneAPIURL        string = "https://flights.ctrip.com/itinerary/api/12808/products"
	APIRequestOrigin   string = "https://flights.ctrip.com"
	APIRequestReferer  string = "https://flights.ctrip.com/itinerary/oneway/bjs-ctu?date=2019-11-15"
	DepartureStrFormat string = "\033[31m(始)\033[0m:%s%s(%s)"
	ArrivalStrFormat   string = "\033[32m(终)\033[0m:%s%s(%s)"
)

var FlightTableHeader = []string{"航空公司", "航班号", "起飞", "起飞时间", "到达", "到达时间", "机型", "餐食", "准点率", "经济舱", "商务舱", "头等舱"}

// 国外航线查询到相关常量
const (
	CityCodeURL                string = "https://flights.ctrip.com/international/search/api/poi/search?"
	FormDataURL                string = "https://flights.ctrip.com/international/search/oneway-{dep}-{arr}?depdate={date}&cabin={cabin}&adult=1&child=0&infant=0"
	OverSeaAirplaneURL         string = "https://flights.ctrip.com/international/search/api/search/batchSearch?v="
	OverSeaAirplanePullDataURL string = "https://flights.ctrip.com/international/search/api/search/pull/{searchId}?v="
)

var CabinNameCode = map[string]string{
	"经济舱":    "y_s",
	"超级经济舱":  "y_s",
	"商务/头等舱": "c_f",
	"商务舱":    "c",
	"公务舱":    "c",
	"头等舱":    "f",
}
var OverSeaFlightTableHeader = []string{"航班号", "航空公司", "机型", "起飞地", "起飞时间", "到达地", "到达时间", "飞行时间", "转机时间"}
var OverSeaFlightTableFooter = []string{"", "", "", "", "", "", "总飞行时长"}

// 机场和航班号信息查询的相关常量
const (
	AirportDepAPIURL   string = "https://adsbapi.variflight.com/adsb/airport/api/departures"
	AirportArrAPIURL   string = "https://adsbapi.variflight.com/adsb/airport/api/arrival"
	FlightNumberAPIURL string = "https://adsbapi.variflight.com/adsb/index/advancedSearch"
)

var FlightNumberInfoTableHeader = []string{"航班状态", "航班号", "出发机场", "到达机场", "计划起飞时间", "实际起飞时间", "计划到达时间", "实际到达时间", "机型", "飞机注册号"}

// TODO: 状态 3 目前不知道是啥
var FlightNumberStatus = map[int64]string{
	0:  "计划",
	1:  "起飞",
	2:  "到达",
	4:  "延误",
	73: "提前取消",
}
var AirportInfoDepTableHeader = []string{"航班号", "机型", "到达地", "到达机场", "计划起飞时间", "实际起飞时间", "状态"}
var AirportInfoArrTableHeader = []string{"航班号", "机型", "出发地", "出发机场", "计划到达时间", "实际到达时间", "状态"}
