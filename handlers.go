package main

import (
	"fmt"
	"github.com/thebotguys/golang-bittrex-api/bittrex"
	"github.com/markcheno/go-talib"
	"time"
	"net/http"
	"strconv"
	"encoding/json"
	"math"
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
	
	candleSticks, err := bittrex.GetTicks("USDT-BTC", "thirtyMin")
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
	if !stringInSlice(indicator, []string{"ema", "wma"}) {
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
		panic(err)
	}
		
	candleSticks, err := bittrex.GetTicks(market, "thirtyMin")
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
	}
	

	var res PlotPoints
	for i, indicatorValue := range indicatorData {
		res = append(res, PlotPoint{time.Time(candleSticks[i].Timestamp).String(), strconv.FormatFloat(indicatorValue, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTraderStart(w http.ResponseWriter, r *http.Request) {
	err := bittrex.IsAPIAlive()
	if err != nil {
		fmt.Println("Can not reach Bittrex API servers: ", err)
	}
	
	candleSticks, err := bittrex.GetTicks("USDT-BTC", "thirtyMin")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
	}

	// listen to ticks
	// spot wma cross
	// buy/sell

	var closes []float64
	for _, candle := range candleSticks {
		closes = append(closes, candle.Close)
	}
	fmt.Println("Trading started at", time.Now().String())
	for {
		select {
			case <-time.After(30 * time.Minute):
				fmt.Println("Tick")
				closes = closes[len(closes)-1000:] //take 1000 values
				candleStick, err := bittrex.GetLatestTick("USDT-BTC", "thirtyMin")
				if err != nil {
					fmt.Println("Latest stick was not received: ", err)
				} else {
					closes = append(closes, candleStick.Close)
				}

				indicatorData1 := talib.Wma(closes, 50)
				indicatorData2 := talib.Wma(closes, 20)
			
				if ((indicatorData1[len(indicatorData1)-1] < indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] > indicatorData2[len(indicatorData1)-2])) ||	((indicatorData1[len(indicatorData1)-1] > indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] < indicatorData2[len(indicatorData1)-2])) {
					action := "sell"
					if indicatorData2[len(indicatorData1)-1] - indicatorData2[len(indicatorData1)-2] > 0 {
						action = "buy"
					}
					marketSummary, err := bittrex.GetMarketSummary("USDT-BTC")
					if err != nil {
						fmt.Println("ERROR OCCURRED: ", err)
					} else {
						fmt.Println(marketSummary)
						fmt.Println("WMA crossed, action:", action, ", price:", marketSummary.Ask)
					}
				} else {
					distance := math.Min(math.Abs(indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1]), math.Abs(indicatorData2[len(indicatorData2)-2] - indicatorData1[len(indicatorData1)-2]))
					fmt.Println("Distance:", distance)
				}
			
		}
	}
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}
//Strategies
// Floor finder
// Pump resolver
	