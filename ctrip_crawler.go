package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/go-resty/resty"
	"github.com/liyu4/tablewriter"
	"github.com/tidwall/gjson"
)

const (
	PlaneAPIURL       string = "https://flights.ctrip.com/itinerary/api/12808/products"
	ContentType       string = "application/json"
	APIRequestOrigin  string = "https://flights.ctrip.com"
	APIRequestReferer string = "https://flights.ctrip.com/itinerary/oneway/bjs-ctu?date=2019-11-15"

	DepartureStrFormat string = "\033[31m(始)\033[0m:%s%s(%s)"
	ArrivalStrFormat   string = "\033[32m(终)\033[0m:%s%s(%s)"

	HasMeal    string = "有餐食"
	HasNotMeal string = "无餐食"

	EconomyClassName  string = "经济舱"
	BusinessClassName string = "商务舱"
	FirstClassName    string = "头等舱"
)

var FlightTableHeader = []string{"航空公司", "航班号", "起飞", "起飞时间", "到达", "到达时间", "机型", "餐食", "准点率", "经济舱", "商务舱", "头等舱"}
var CabinClassMap = map[string]string{
	"Y": EconomyClassName,
	"C": BusinessClassName,
	"F": FirstClassName,
}

type AirportParams struct {
	ACity     string `json:"acity"`
	ACityName string `json:"acityname"`
	Date      string `json:"date"`
	DCity     string `json:"dcity"`
	DCityName string `json:"dcityname"`
}

type FlightTablePayload struct {
	APParams    []AirportParams `json:"airportParams"`
	Army        bool            `json:"army"`
	ClassType   string          `json:"classType"`
	FlightWay   string          `json:"flightWay"`
	HasBaby     bool            `json:"hasBaby"`
	HasChild    bool            `json:"hasChild"`
	Params      []AirportParams `json:"params"`
	SearchIndex int             `json:"searchIndex"`
}

type CabinData struct {
	CabinType      string
	CabinPriceRate float64
	CabinRestSeats int64
}

type CtripCrawler struct {
	RestClient *resty.Client

	IsOnlyLowerPrice bool
	FlightTable      *tablewriter.Table
}

func NewCtripCrawler() *CtripCrawler {
	ctrip := &CtripCrawler{}
	ctrip.initCtripCrawler()
	return ctrip
}

// 初始化
func (c *CtripCrawler) initCtripCrawler() {
	c.RestClient = resty.New()
}

// 初始化表格
func (c *CtripCrawler) initFlightTable() {
	flightTable := tablewriter.NewColorWriter(os.Stdout)
	flightTable.SetAlignment(tablewriter.ALIGN_LEFT)
	flightTable.SetHeader(FlightTableHeader)
	c.FlightTable = flightTable
}

// 构造请求参数
func (c *CtripCrawler) getFlightTablePayload(departureCityName, arriveCityName, date, classType, tripType string) string {
	airportParams := AirportParams{
		ACity:     cityNameCode[arriveCityName],
		ACityName: arriveCityName,
		Date:      date,
		DCity:     cityNameCode[departureCityName],
		DCityName: departureCityName,
	}
	payload := FlightTablePayload{
		APParams:    make([]AirportParams, 0),
		Army:        false,
		ClassType:   classType,
		FlightWay:   tripType,
		HasBaby:     false,
		HasChild:    false,
		SearchIndex: 1,
	}
	payload.APParams = append(payload.APParams, airportParams)
	payload.Params = append(payload.Params, airportParams)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("[Flight-Go]Json 转换出错!")
	}
	return string(jsonData)
}

// 表格查询
func (c *CtripCrawler) runFlightTableCrawler(departureCityName, arriveCityName, date, classType, tripType string, onlyLowPrice bool) {
	c.IsOnlyLowerPrice = onlyLowPrice
	payloadData := c.getFlightTablePayload(departureCityName, arriveCityName, date, classType, tripType)
	dataResp, err := c.RestClient.R().
		SetHeader("content-type", ContentType).
		SetHeader("origin", APIRequestOrigin).
		SetHeader("referer", APIRequestReferer).
		SetHeader("user-agent", UserAgent).
		SetBody(payloadData).
		Post(PlaneAPIURL)
	if err != nil {
		logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
	} else {
		if dataResp.String() != "" {
			c.parseFlightTable(gjson.Parse(dataResp.String()))
		} else {
			logger.Error("[Flight-Go]接口数据为空!")
		}
	}
}

// 解析表格
func (c *CtripCrawler) parseFlightTable(tableJson gjson.Result) {
	c.initFlightTable()
	// 航班数据
	flightRouteList := tableJson.Get("data").Get("routeList").Array()
	for _, flightInfoHeader := range flightRouteList {
		// 判断线路类型 Flight 飞行；FlightTrain 空地联运
		tripType := flightInfoHeader.Get("routeType").String()
		if tripType == "Flight" {
			flightLegs := flightInfoHeader.Get("legs").Array()
			for _, flightInfo := range flightLegs {
				// 核心信息
				flightData := flightInfo.Get("flight")
				// 航空公司和航班号
				airlineName := flightData.Get("airlineName").String()
				flightNumber := flightData.Get("flightNumber").String()
				// 起飞
				departureCityName := flightData.Get("departureAirportInfo").Get("cityName").String()
				departureAirport := flightData.Get("departureAirportInfo").Get("airportName").String()
				departureAirportTerminalName := flightData.Get("departureAirportInfo").Get("terminal").Get("name").String()
				departureInfo := fmt.Sprintf(DepartureStrFormat, departureCityName, departureAirport, departureAirportTerminalName)
				dTime, _ := time.Parse("2006-01-02 15:04:05", flightData.Get("departureDate").String())
				departureTime := dTime.Format("15:04")
				// 到达
				arrivalCityName := flightData.Get("arrivalAirportInfo").Get("cityName").String()
				arrivalAirport := flightData.Get("arrivalAirportInfo").Get("airportName").String()
				arrivalAirportTerminalName := flightData.Get("arrivalAirportInfo").Get("terminal").Get("name").String()
				arrivalInfo := fmt.Sprintf(ArrivalStrFormat, arrivalCityName, arrivalAirport, arrivalAirportTerminalName)
				aTime, _ := time.Parse("2006-01-02 15:04:05", flightData.Get("arrivalDate").String())
				arrivalTime := aTime.Format("15:04")
				// 机型
				aircraftTypeName := flightData.Get("craftTypeName").String()
				aircraftTypeName = strings.Replace(aircraftTypeName, "全新", "", -1)
				aircraftTypeName = strings.Replace(aircraftTypeName, " ", "", -1)
				// TODO 暂时替换(嫌他太长了)
				aircraftTypeName = strings.Replace(aircraftTypeName, "A350-900", "350", -1)
				aircraftTypeCode := flightData.Get("craftTypeCode").String()
				aircraftInfo := fmt.Sprintf("%s(%s)", aircraftTypeName, aircraftTypeCode)
				// 餐食
				mealFlag := flightData.Get("mealFlag").Bool()
				var mealFlagStr = HasNotMeal
				if mealFlag {
					mealFlagStr = HasMeal
				}
				mealInfo := mealFlagStr
				// 准点率
				punctualityRate := fmt.Sprintf("%s", flightData.Get("punctualityRate").String())
				// 航班价格
				cabinPricesFunc := func(cabins []gjson.Result) (economyClassPrices, businessClassPrices, firstClassPrices []string) {
					cabinPriceMap := make(map[int]CabinData)
					for _, cabinInfo := range cabins {
						cabinPrice := cabinInfo.Get("price").Get("price").Int()
						// TODO 不同价格, 暂时不清楚有啥用
						//cabinPrice := cabinInfo.Get("price").Get("salePrice").Int()
						//cabinPrice := cabinInfo.Get("price").Get("printPrice").Int()
						cabinStruct := CabinData{
							CabinType:      CabinClassMap[cabinInfo.Get("cabinClass").String()],
							CabinPriceRate: cabinInfo.Get("price").Get("rate").Float(),
							CabinRestSeats: cabinInfo.Get("seatCount").Int(),
						}
						price := *(*int)(unsafe.Pointer(&cabinPrice))
						cabinPriceMap[price] = cabinStruct
					}
					// 排序
					var keys []int
					for p := range cabinPriceMap {
						keys = append(keys, p)
					}
					sort.Ints(keys)
					economyClassPrices = make([]string, 0)
					businessClassPrices = make([]string, 0)
					firstClassPrices = make([]string, 0)
					for _, price := range keys {
						cabinType := cabinPriceMap[price].CabinType
						// 折扣信息
						var rates string
						if cabinPriceMap[price].CabinPriceRate == 1.0 {
							rates = fmt.Sprintf("无折扣")
						} else {
							rates = fmt.Sprintf("%.1f折", cabinPriceMap[price].CabinPriceRate*10)
						}
						switch cabinType {
						case EconomyClassName:
							economyClassPrices = append(economyClassPrices, fmt.Sprintf("价格:%d元（%s,剩余:%d张）", price, rates, cabinPriceMap[price].CabinRestSeats))
						case BusinessClassName:
							businessClassPrices = append(businessClassPrices, fmt.Sprintf("价格:%d元（%s,剩余:%d张）", price, rates, cabinPriceMap[price].CabinRestSeats))
						case FirstClassName:
							firstClassPrices = append(firstClassPrices, fmt.Sprintf("价格:%d元（%s,剩余:%d张）", price, rates, cabinPriceMap[price].CabinRestSeats))
						default:
							continue
						}
					}
					return economyClassPrices, businessClassPrices, firstClassPrices
				}
				var economyClassPrice = "无"
				var businessClassPrice = "无"
				var firstClassPrice = "无"
				economyClassPrices, businessClassPrices, firstClassPrices := cabinPricesFunc(flightInfo.Get("cabins").Array())
				if c.IsOnlyLowerPrice {
					if len(economyClassPrices) > 0 {
						economyClassPrice = economyClassPrices[0]
					}
					if len(businessClassPrices) > 0 {
						businessClassPrice = businessClassPrices[0]
					}
					if len(firstClassPrices) > 0 {
						firstClassPrice = firstClassPrices[0]
					}
				} else {
					// TODO 多个价格输出(不知道怎么展示会比较好), 后面再想办法, 暂时再沿用上面的方法进行输出
					if len(economyClassPrices) > 0 {
						economyClassPrice = economyClassPrices[0]
					}
					if len(businessClassPrices) > 0 {
						businessClassPrice = businessClassPrices[0]
					}
					if len(firstClassPrices) > 0 {
						firstClassPrice = firstClassPrices[0]
					}
				}
				// 合并到表格中
				row := []string{
					airlineName,
					flightNumber,
					departureInfo,
					departureTime,
					arrivalInfo,
					arrivalTime,
					aircraftInfo,
					mealInfo,
					punctualityRate,
					economyClassPrice,
					businessClassPrice,
					firstClassPrice,
				}
				c.FlightTable.Append(row)
			}
		}
	}
	c.FlightTable.Render()
}
