package main

import (
	"github.com/go-resty/resty"
	"github.com/liyu4/tablewriter"
	"github.com/tidwall/gjson"
	"os"
)

const (
	AirportDepAPIURL           string = "https://adsbapi.variflight.com/adsb/airport/api/departures"
	AirportArrAPIURL           string = "https://adsbapi.variflight.com/adsb/airport/api/arrival"
	FlightNumberAPIURL         string = "https://adsbapi.variflight.com/adsb/index/advancedSearch"
	FlightNumberAPIContentType string = "application/x-www-form-urlencoded"
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

type VariFlightCrawler struct {
	RestClient *resty.Client

	FlightNumberInfoTable *tablewriter.Table
	AirportInfoTable      *tablewriter.Table
}

func NewVariFlightCrawler() *VariFlightCrawler {
	vari := &VariFlightCrawler{}
	vari.initVariFlightCrawler()
	return vari
}

// 初始化
func (v *VariFlightCrawler) initVariFlightCrawler() {
	v.RestClient = resty.New()
}

// 构造航班信息请求数据
func (v *VariFlightCrawler) getFlightNumberPayload(flightNumber, date string) (payload map[string]string) {
	payload = make(map[string]string)
	payload["searchText"] = flightNumber
	payload["searchDate"] = date
	payload["timeZone"] = "-28800"
	return
}

// 初始化航班信息表格
func (v *VariFlightCrawler) initFlightNumberInfoTable() {
	table := tablewriter.NewColorWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(FlightNumberInfoTableHeader)
	v.FlightNumberInfoTable = table
}

// 解析信息成表格
func (v *VariFlightCrawler) parseDataToTable(tableJson gjson.Result) {
	v.initFlightNumberInfoTable()
	flightData := tableJson.Get("data").Array()
	for _, data := range flightData {
		flightStatus := FlightNumberStatus[data.Get("flightStatusCode").Int()]
		flightNumber := data.Get("fnum").String()
		departureAirportName := data.Get("forgAptCname").String()
		arrivalAirportName := data.Get("fdstAptCname").String()
		estimatedDepTime := timestampToTime(data.Get("scheduledDeptime").Int())
		actualDepTime := timestampToTime(data.Get("actualDeptime").Int())
		estimatedArrTime := timestampToTime(data.Get("scheduledArrtime").Int())
		actualArrTime := timestampToTime(data.Get("actualArrtime").Int())
		airCraftType := data.Get("ftype").String()
		airCraftNumber := data.Get("aircraftNumber").String()
		// 每一行的数据
		row := []string{
			flightStatus,
			flightNumber,
			departureAirportName,
			arrivalAirportName,
			estimatedDepTime,
			actualDepTime,
			estimatedArrTime,
			actualArrTime,
			airCraftType,
			airCraftNumber,
		}
		v.FlightNumberInfoTable.Append(row)
	}
	v.FlightNumberInfoTable.Render()
}

// 查询航班信息
func (v *VariFlightCrawler) runFlightInfo(flightNumber, date string) {
	payloadData := v.getFlightNumberPayload(flightNumber, date)
	dataResp, err := v.RestClient.R().
		SetQueryParam("lang", "zh_CN").
		SetHeader("Content-Type", FlightNumberAPIContentType).
		SetHeader("User-Agent", UserAgent).
		SetFormData(payloadData).
		Post(FlightNumberAPIURL)
	if err != nil {
		logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
	} else {
		if dataResp.String() != "" {
			v.parseDataToTable(gjson.Parse(dataResp.String()))
		} else {
			logger.Error("[Flight-Go]接口数据为空!")
		}
	}
}

// 初始化机场进出港表格
func (v *VariFlightCrawler) initAirportInfoTable() {
	table := tablewriter.NewColorWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	v.AirportInfoTable = table
}

// 解析进场进出港数据表格
func (v *VariFlightCrawler) parseAirportInfoTable(depOrArr string, tableJson gjson.Result) {
	airportInfoList := tableJson.Get("list").Array()

	for _, airportInfo := range airportInfoList {
		flightNumber := airportInfo.Get("fnum").String()
		airCraftType := airportInfo.Get("ftype").String()
		flightStatus := FlightNumberStatus[airportInfo.Get("flightStatusCode").Int()]
		if depOrArr == "dep" {
			destinationName := airportInfo.Get("fdstAptCcity").String()
			destinationAirportName := airportInfo.Get("fdstAptCname").String()
			scheduleDepTime := timestampToTime(airportInfo.Get("scheduledDeptime").Int())
			actualDepTime := timestampToTime(airportInfo.Get("estimatedDeptime").Int())
			row := []string{
				flightNumber,
				airCraftType,
				destinationName,
				destinationAirportName,
				scheduleDepTime,
				actualDepTime,
				flightStatus,
			}
			v.AirportInfoTable.Append(row)
		} else if depOrArr == "arr" {
			destinationName := airportInfo.Get("forgAptCcity").String()
			destinationAirportName := airportInfo.Get("forgAptCname").String()
			scheduledArrTime := timestampToTime(airportInfo.Get("scheduledArrtime").Int())
			var actualArrTimeStr string
			actualArrTime := airportInfo.Get("actualArrtime").Int()
			if actualArrTime == 0 {
				actualArrTimeStr = timestampToTime(airportInfo.Get("estimatedArrtime").Int())
			} else {
				actualArrTimeStr = timestampToTime(actualArrTime)
			}
			row := []string{
				flightNumber,
				airCraftType,
				destinationName,
				destinationAirportName,
				scheduledArrTime,
				actualArrTimeStr,
				flightStatus,
			}
			v.AirportInfoTable.Append(row)
		}
	}
	v.AirportInfoTable.Render()
}

// 查询机场进出港信息
func (v *VariFlightCrawler) runAirportInfo(areaName, depOrArr string) {
	v.initAirportInfoTable()
	var ReqURL string
	switch depOrArr {
	case "dep":
		ReqURL = AirportDepAPIURL
		v.AirportInfoTable.SetHeader(AirportInfoDepTableHeader)
	case "arr":
		ReqURL = AirportArrAPIURL
		v.AirportInfoTable.SetHeader(AirportInfoArrTableHeader)
	}
	dataResp, err := v.RestClient.R().
		SetQueryParam("lang", "zh_CN").
		SetQueryParam("iata", cityNameCode[areaName]).
		SetQueryParam("pageSize", "15").
		SetQueryParam("pageNum", "1").
		SetHeader("User-Agent", UserAgent).
		Get(ReqURL)
	if err != nil {
		logger.Fatalf("[Flight-Go]接口请求出错!, 错误原因: %s", err.Error())
	} else {
		if dataResp.String() != "" {
			v.parseAirportInfoTable(depOrArr, gjson.Parse(dataResp.String()))
		} else {
			logger.Error("[Flight-Go]接口数据为空!")
		}
	}
}
