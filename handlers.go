package main

import (
	"fmt"
	"github.com/thebotguys/golang-bittrex-api/bittrex"
	"github.com/markcheno/go-talib"
	"time"
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/spf13/viper"
	"os"
	"io/ioutil"
)

type PlotPoint struct {
	Date string
	Value string
}
type PlotPoints []PlotPoint




func handleChartBtcUsd(w http.ResponseWriter, r *http.Request) {
	
	err := bittrex.IsAPIAlive()
	if err != nil {
		fmt.Println("Can not reach Bittrex API servers: ", err)
	}
	
	candleSticks, err := bittrex.GetTicks("USDT-BTC", "fiveMin")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
	}
	fmt.Println("Ticks collected: ", len(candleSticks))
	
	var res PlotPoints
	for _, candle := range candleSticks {
		res = append(res, PlotPoint{time.Time(candle.Timestamp).String(), strconv.FormatFloat(candle.Close, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleEmaBtcUsd(w http.ResponseWriter, r *http.Request) {
	
	err := bittrex.IsAPIAlive()
	if err != nil {
		fmt.Println("Can not reach Bittrex API servers: ", err)
	}
	
	candleSticks, err := bittrex.GetTicks("USDT-BTC", "thirtyMin")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
	}
	
	var closes []float64
	for _, candle := range candleSticks {
		closes = append(closes, candle.Close)
	}
	
	interval, err := strconv.Atoi(r.URL.Query().Get("interval"))
	if err != nil || interval < 5  {
		interval = 5
	}
	fmt.Println("Getting EMA for USDT-BTC (", interval, ")")
	emaData := talib.Ema(closes, interval)

	var res PlotPoints
	for i, emaValue := range emaData {
		res = append(res, PlotPoint{time.Time(candleSticks[i].Timestamp).String(), strconv.FormatFloat(emaValue, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleIndicatorChart(w http.ResponseWriter, r *http.Request) {
	indicator := r.URL.Query().Get("name")
	if !stringInSlice(indicator, []string{"ema", "wma", "trima", "rsi", "httrendline"}) {
		panic("indicator is not recognized")
	}
	market := r.URL.Query().Get("market") //"USDT-BTC"
	interval, err := strconv.Atoi(r.URL.Query().Get("interval"))
	if err != nil || interval < 5  {
		interval = 5
	}
	
	err = bittrex.IsAPIAlive()
	if err != nil {
		fmt.Println("Can not reach Bittrex API servers: ", err)
		// panic(err)
	}
		
	candleSticks, err := bittrex.GetTicks(market, "fiveMin")
	if err != nil {
		panic(err)
	}
	
	var closes []float64
	for _, candle := range candleSticks {
		closes = append(closes, candle.Close)
	}	
	
	
	var indicatorData []float64

	fmt.Println("Indicator: ", indicator, market, interval)
	
	switch indicator {
	case "ema": indicatorData = talib.Ema(closes, interval)
	case "wma": indicatorData = talib.Wma(closes, interval)
	case "trima": indicatorData = talib.Trima(closes, interval)
	case "rsi": indicatorData = talib.Rsi(closes, interval)
	case "httrendline": indicatorData = talib.HtTrendline(closes)
	}
	

	var res PlotPoints
	for i, indicatorValue := range indicatorData {
		res = append(res, PlotPoint{time.Time(candleSticks[i].Timestamp).String(), strconv.FormatFloat(indicatorValue, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTraderStart(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	strategy := r.URL.Query().Get("strategy")

	err := bittrex.IsAPIAlive()
	if err != nil {
		fmt.Println("Can not reach Bittrex API servers: ", err)
		//panic(err)
	}
	
	viper.SetConfigType("json")
	file, err := os.Open("config/trading.json")
	if err != nil { panic("Config file does not exist.") }	
	viper.ReadConfig(file)

	config := viper.GetStringMapString("strategies." + strategy)

	switch strategy {
	case "wma": go Begin(market, config, strategyWma)
	case "dip": go Begin(market, config, strategyDip)
	default: panic("Unrecognized strategy.")
	}
	
	jsonResponse, _ := json.Marshal("OK")
	fmt.Fprintf(w, string(jsonResponse))
}

func handleStrategyTest(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	
	viper.SetConfigType("json")
	file, err := os.Open("config/config.json")
	if err != nil { panic("Config file does not exist.") }	
	viper.ReadConfig(file)

	config := viper.GetStringMapString("exchanges.bittrex")


	if config["connection_check"] == "Y" {
		err := bittrex.IsAPIAlive()
		if err != nil {
			fmt.Println("Can not reach Bittrex API servers: ", err)
			panic(err)
		}
	}
	
	// get data
	var candleSticks bittrex.CandleSticks
	if false {
		candleSticks, err = bittrex.GetTicks(market, "fiveMin")
		if err != nil {
			fmt.Println("ERROR OCCURRED: ", err)
			panic(err)
		}

		// dump to a file
		go func(cs *bittrex.CandleSticks) {
			jsonResponse, _ := json.Marshal(cs)
			fmt.Fprintf(w, string(jsonResponse))
			err := ioutil.WriteFile("./testbeds/tb1.json", jsonResponse, 0644)
			if err != nil {
				fmt.Println(err)
			}
		}(&candleSticks)

	} else {
		dat, err := ioutil.ReadFile("./testbeds/tb1.json")
		if err != nil {
			fmt.Println("ERROR OCCURRED: ", err)
			panic(err)
		}
		fmt.Print(string(dat))
		err = json.Unmarshal(dat, &candleSticks)
		fmt.Println(err)
	}
	
	fmt.Print(candleSticks)
		

	// test through it
	var result TestingResult
	var lastPrice float64 = 0
	var bottomLine float64 = 0
	var lastAction MarketAction

	for i := 0; i <= len(candleSticks) - 1000; i++ {
		t := candleSticks[0:1000+i]
		marketAction := strategyDip(market, &t, &lastAction)
		if (marketAction != nil) {
			result.Actions = append(result.Actions, TestingResultAction{marketAction.Action, marketAction.Market, marketAction.Price, marketAction.Time.String()})
			if (marketAction.Action == MarketActionBuy) {
				lastPrice = marketAction.Price
			} else if marketAction.Action == MarketActionSell {
				if lastPrice > 0 {
					bottomLine = bottomLine + marketAction.Price - lastPrice
					lastPrice = 0
					result.Balances = append(result.Balances, PlotPoint{time.Time(marketAction.Time).String(), strconv.FormatFloat(bottomLine, 'f', 6, 64)})
				}
			}
		}
	}

	result.FinalBalance = bottomLine
	jsonResponse, _ := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResponse))
}
//Strategies
// Floor finder
// Pump resolver
	