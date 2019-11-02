package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/go-resty/resty"
	"github.com/liyu4/tablewriter"
	"github.com/tidwall/gjson"
)

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
				// TODO 暂时替换(嫌他太长了) 貌似原数据的是 <全新 A350-900>
				aircraftTypeName = strings.Replace(aircraftTypeName, "全新", "", -1)
				aircraftTypeName = strings.Replace(aircraftTypeName, " ", "", -1)
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
					airlineName, flightNumber, departureInfo, departureTime, arrivalInfo, arrivalTime,
					aircraftInfo, mealInfo, punctualityRate, economyClassPrice, businessClassPrice, firstClassPrice,
				}
				c.FlightTable.Append(row)
			}
		}
	}
	c.FlightTable.Render()
}

// 国内航班查询
func (c *CtripCrawler) runMainLandFlightTableCrawler(departureCityName, arriveCityName, date, tripType string, onlyLowPrice bool) {
	c.IsOnlyLowerPrice = onlyLowPrice
	payloadData := c.getFlightTablePayload(departureCityName, arriveCityName, date, "ALL", tripType)
	dataResp, err := c.RestClient.R().
		SetHeader("content-type", ContentTypeJson).
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
1、获取 cityCode 的接口地址: https://flights.ctrip.com/international/search/api/poi/search?key=
2、表单数据接口：https://flights.ctrip.com/international/search/oneway-{dep}-{arr}?depdate={date}&cabin=y_s&adult=1&child=0&infant=0
3、加密参数
加密字段在 Header 中: sign
Js 加密参数方法源文件: https://webresource.c-ctrip.com/ResFltIntlOnline/R11/assets/list.js?v=20191031:formatted
源代码:
{
	key: "genAntiCrawlerHeader",
	value: function(e) {
		var t = "";
		return e.get("flightSegments").valueSeq().forEach(function(e) {
			var n = e.get("departureCityCode")
			  , r = e.get("arrivalCityCode")
			  , i = e.get("departureDate");
			t += n + r + i
		}),
		{
			sign: (new b.a).update(e.get("transactionID") + t).digest("hex")
		}
	}
}
*/
// 通过国家或者城市名查询城市号
func (c *CtripCrawler) getCityCode(cityName string) string {
	params := url.Values{}
	params.Add("key", cityName)
	dataResp, err := c.RestClient.R().
		SetHeader("Accept", ContentTypeJson).
		SetHeader("user-agent", UserAgent).
		Get(fmt.Sprintf("%s%s", CityCodeURL, params.Encode()))
	if err != nil {
		logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
	} else {
		if dataResp.String() != "" {
			DataArray := gjson.Parse(dataResp.String()).Get("Data").Array()
			if len(DataArray) > 0 {
				return DataArray[0].Get("Code").String()
			} else {
				return ""
			}
		} else {
			logger.Error("[Flight-Go]接口数据为空!")
			return ""
		}
	}
	return ""
}

// 获取 form 表单数据
func (c *CtripCrawler) getAPIFormData(departureCityName, arriveCityName, date, cabin string) string {
	depCode := c.getCityCode(departureCityName)
	arrCode := c.getCityCode(arriveCityName)
	reqURL := stringFormat(FormDataURL, "{dep}", depCode, "{arr}", arrCode, "{date}", date, "{cabin}", cabin)
	dataResp, err := c.RestClient.R().SetHeader("User-Agent", UserAgent).Get(reqURL)
	if err != nil {
		logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
	} else {
		formDataReg := regexp.MustCompile("GlobalSearchCriteria =(.*?);")
		formData := formDataReg.FindStringSubmatch(dataResp.String())
		if len(formData) > 1 {
			return formData[1]
		} else {
			return ""
		}
	}
	return ""
}

// 生成加密参数 sign
func (c *CtripCrawler) generateSignValue(data string) (string, string) {
	jsonData := gjson.Parse(data)
	transactionId := jsonData.Get("transactionID").String()
	depCode := jsonData.Get("flightSegments").Array()[0].Get("departureCityCode").String()
	arrCode := jsonData.Get("flightSegments").Array()[0].Get("arrivalCityCode").String()
	date := jsonData.Get("flightSegments").Array()[0].Get("departureDate").String()
	return transactionId, getRandomMD5ByCustomStr(fmt.Sprintf("%s%s%s%s", transactionId, depCode, arrCode, date))
}

// 解析国外航班数据表格
func (c *CtripCrawler) parseOverSeaFlightTable(tableJson []gjson.Result, cabinName string) {
	for _, flightData := range tableJson {
		// 给机票表格
		eachFlightTable := tablewriter.NewColorWriter(os.Stdout)
		eachFlightTable.SetAlignment(tablewriter.ALIGN_LEFT)
		eachFlightTable.SetHeader(OverSeaFlightTableHeader)
		// 机票航段信息
		flightSegments := flightData.Get("flightSegments").Array()[0]
		// 各段航班信息
		for _, flightInfo := range flightSegments.Get("flightList").Array() {
			// 航班号
			flightCode := flightInfo.Get("flightNo").String()
			// 航空公司
			airlineName := flightInfo.Get("marketAirlineName").String()
			// 机型
			airCraftType := flightInfo.Get("aircraftName").String()
			// 起飞地
			departureCityName := fmt.Sprintf("%s-%s-%s(%s)", flightInfo.Get("departureCountryName").String(), flightInfo.Get("departureCityName").String(), flightInfo.Get("departureAirportName").String(), flightInfo.Get("departureTerminal").String())
			// 起飞时间
			departureTime := flightInfo.Get("departureDateTime").String()
			// 到达地
			arrivalCityName := fmt.Sprintf("%s-%s-%s(%s)", flightInfo.Get("arrivalCountryName").String(), flightInfo.Get("arrivalCityName").String(), flightInfo.Get("arrivalAirportName").String(), flightInfo.Get("arrivalTerminal").String())
			// 到达时间
			arrivalTime := flightInfo.Get("arrivalDateTime").String()
			// 飞行时间
			flightHour, flightMinutes := minutesToHour(flightInfo.Get("duration").Int())
			flightTime := fmt.Sprintf("%d 小时 %d 分钟", flightHour, flightMinutes)
			// 转机时间
			transferHour, transferMinutes := minutesToHour(flightInfo.Get("transferDuration").Int())
			transferTime := fmt.Sprintf("%d 小时 %d 分钟", transferHour, transferMinutes)
			if transferHour == 0 && transferMinutes == 0 {
				transferTime = "-"
			}
			// 写入表格数据
			row := []string{
				flightCode, airlineName, airCraftType, departureCityName, departureTime,
				arrivalCityName, arrivalTime, flightTime, transferTime,
			}
			eachFlightTable.Append(row)
		}
		// 飞行时间
		hour, minutes := minutesToHour(flightSegments.Get("duration").Int())
		totalFlightTime := fmt.Sprintf("%d 小时 %d 分钟", hour, minutes)
		// 将机票价格展示（目前暂时没有加入）
		flightPricesFunc := func() []int64 {
			var priceList []int64
			for _, price := range flightData.Get("priceList").Array() {
				totalPrice := price.Get("adultPrice").Int() + price.Get("adultTax").Int()
				priceList = append(priceList, totalPrice)
			}
			return priceList
		}
		OverSeaFlightTableFooter[3] = fmt.Sprintf("当前舱位: %s", cabinName)
		OverSeaFlightTableFooter[4] = fmt.Sprintf("最低价格: %d 元", flightPricesFunc()[0])
		// 渲染表格
		eachFlightTable.SetFooter(append(OverSeaFlightTableFooter, totalFlightTime, ""))
		eachFlightTable.Render()
	}
}

// 国外航班舱位信息
func (c *CtripCrawler) overSeaFlightSeatTypeToCabinName(seatType string) string {
	cabinName := seatType
	if seatType == "" {
		cabinName = "y_s"
	} else {
		cabinName = CabinNameCode[seatType]
	}
	if cabinName == "" {
		logger.Fatal("[Flight-Go]舱位参数错误!")
	}
	return cabinName
}

// 国外航班查询
func (c *CtripCrawler) runOverSeaFlightTableCrawler(departureCityName, arriveCityName, date, seatType string) {
	cabinName := c.overSeaFlightSeatTypeToCabinName(seatType)
	body := c.getAPIFormData(departureCityName, arriveCityName, date, cabinName)
	if body == "" {
		logger.Fatal("[Flight-Go]接口请求出错! 数据异常!")
	}
	transactionId, sign := c.generateSignValue(body)
	// 获取航班数据
	reqURL := OverSeaAirplaneURL
	var allFlightData []gjson.Result
	for {
		//logger.Infof("[Flight-Go]当前请求的地址: %s", reqURL)
		dataResp, err := c.RestClient.R().
			SetHeader("Content-Type", ContentTypeJson).
			SetHeader("User-Agent", UserAgent).
			SetHeader("sign", sign).
			SetHeader("transactionid", transactionId).
			SetBody(body).
			Post(reqURL)
		if err != nil {
			logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
		} else {
			if dataResp.String() != "" {
				respJsonData := gjson.Parse(dataResp.String())
				isFullLoadData := respJsonData.Get("data").Get("context").Get("finished").Bool()
				allFlightData = append(append(allFlightData), respJsonData.Get("data").Get("flightItineraryList").Array()...)
				if isFullLoadData {
					// 是否加载完全部
					break
				} else {
					reqURL = stringFormat(OverSeaAirplanePullDataURL, "{searchId}", respJsonData.Get("data").Get("context").Get("searchId").String())
				}
			} else {
				logger.Fatal("[Flight-Go]接口请求出错! 数据异常!")
			}
		}
		time.Sleep(time.Second * 1)
	}
	c.parseOverSeaFlightTable(allFlightData, seatType)
}
