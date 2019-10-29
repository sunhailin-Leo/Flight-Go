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
	flightCommands = []*FlightCommand{
		flightTableCommand,
	}
	flightDepartureCityName string
	flightArrivalCityName   string
	flightDate              string
	flightClassType         string
	flightTripType          string
)

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
}
