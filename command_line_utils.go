package main

import "flag"

type FlightCommand struct {
	UsageLine string
	Run       func(args []string) int
	Flag      flag.FlagSet
}

func (c *FlightCommand) Name() string {
	return c.UsageLine
}

var (
	flightTableCommand = &FlightCommand{
		UsageLine: "schedule",
	}
	flightDepartureCityName string
	flightArrivalCityName   string
	flightDate              string
	flightClassType         string
	flightTripType          string
)

var (
	flightNumberInfoCommand = &FlightCommand{
		UsageLine: "code",
	}
	flightNumber          string
	flightNumberCheckDate string
)

var (
	airportInfoCommand = &FlightCommand{
		UsageLine: "airport",
	}
	airportName     string
	airportDepOrArr string
)

var flightCommands = []*FlightCommand{
	flightTableCommand,
	flightNumberInfoCommand,
	airportInfoCommand,
}

func executeAirportInfoTableFunc(args []string) int {
	airportInfoTable := NewVariFlightCrawler()
	airportInfoTable.runAirportInfo(args[0], args[1])
	return 1
}

func executeFlightNumberInfoTableFunc(args []string) int {
	flightNumberTable := NewVariFlightCrawler()
	flightNumberTable.runFlightInfo(args[0], args[1])
	return 1
}

func executeFlightTableFunc(args []string) int {
	flightTable := NewCtripCrawler()
	flightTable.runFlightTableCrawler(args[0], args[1], args[2], "ALL", "Oneway", true)
	return 1
}

func commandLineInit() {
	flightTableCommand.Run = executeFlightTableFunc
	flightTableCommand.Flag.StringVar(&flightDepartureCityName, "dep", "", "需要查询的始发地")
	flightTableCommand.Flag.StringVar(&flightArrivalCityName, "arr", "", "需要查询的目的地")
	flightTableCommand.Flag.StringVar(&flightDate, "date", "", "需要搜索的日期（格式: YYYY-MM-DD 例如: 2019-10-17）")

	flightNumberInfoCommand.Run = executeFlightNumberInfoTableFunc
	flightNumberInfoCommand.Flag.StringVar(&flightNumber, "flightNumber", "", "需要查询的航班号")
	flightNumberInfoCommand.Flag.StringVar(&flightNumberCheckDate, "date", "", "需要搜索的日期（格式: YYYYMMDD 例如: 20191017）")

	airportInfoCommand.Run = executeAirportInfoTableFunc
	airportInfoCommand.Flag.StringVar(&airportName, "airportName", "", "需要查询机场名称（例如: 广州）")
	airportInfoCommand.Flag.StringVar(&airportDepOrArr, "depOrArr", "", "进场的进出港类别")
}
