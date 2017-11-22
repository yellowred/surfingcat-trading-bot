package main

import (
	"fmt"
	"github.com/yellowred/golang-bittrex-api/bittrex"
	"github.com/markcheno/go-talib"
	"time"
	"net/http"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	trading "github.com/yellowred/surfingcat-trading-bot/server/trading"
	message "github.com/yellowred/surfingcat-trading-bot/server/message"
	"github.com/yellowred/surfingcat-trading-bot/server/utils"
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
	if !utils.StringInSlice(indicator, []string{"ema", "wma", "trima", "rsi", "httrendline"}) {
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
	
	bot := trading.NewBot(market, strategy, configManager.StrategyConfig(strategy), exchange.ExchangeClient(exchange.EXCHANGE_PROVIDER_BITTREX, configManager.ExchangeConfig(exchange.EXCHANGE_PROVIDER_BITTREX)))
	go bot.Start()
	jsonResponse, _ := json.Marshal(bot.Uuid)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTraderStop(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	err := message.StopTrader(uuid)
	if err != nil {
		jsonResponse, _ := json.Marshal(err)
		fmt.Fprintf(w, string(jsonResponse))
	} else {
		jsonResponse, _ := json.Marshal(uuid)
		fmt.Fprintf(w, string(jsonResponse))
	}
}


/*
func handleTraderList(w http.ResponseWriter, r *http.Request) {

	for uuid, ch := range traders {

	}
	if traderCh, ok := traders[uuid]; ok {
		traderCh <- ServerMessage{uuid, ServerMessageActionStop}
		close(traderCh)
	}
	jsonResponse, _ := json.Marshal(uuid)
	fmt.Fprintf(w, string(jsonResponse))
}
*/

func handleTraderCheck(w http.ResponseWriter, r *http.Request) {
	client := exchange.ExchangeClient(exchange.EXCHANGE_PROVIDER_BITTREX, configManager.ExchangeConfig(exchange.EXCHANGE_PROVIDER_BITTREX))
	// uuid, err := client.Buy("USDT-BTC", 0.001, 6000)
	// uuid, err := client.Sell("BTC-FCT", 1, 0.01)
	m, err := client.MarketSummary("USDT-BTC")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
	}
	jsonResponse, _ := json.Marshal(m)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTraderBalance(w http.ResponseWriter, r *http.Request) {
	client := exchange.ExchangeClient(exchange.EXCHANGE_PROVIDER_BITTREX, configManager.ExchangeConfig(exchange.EXCHANGE_PROVIDER_BITTREX))
	m, err := client.Balances()
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
	}
	jsonResponse, _ := json.Marshal(m)
	fmt.Fprintf(w, string(jsonResponse))
}


type TestingResult struct {
	Actions []exchange.TestMarketAction
	Balances []exchange.Balance
}

func handleStrategyTest(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	strategy := r.URL.Query().Get("strategy")
	

	btx := exchange.ExchangeClient(exchange.EXCHANGE_PROVIDER_BITTREX, configManager.ExchangeConfig(exchange.EXCHANGE_PROVIDER_BITTREX))
	
	// get data
	var candleSticks []exchange.CandleStick
	var err error
	if false {
		candleSticks, err = btx.AllCandleSticks(market, "fiveMin")
		if err != nil {
			fmt.Println("ERROR OCCURRED: ", err)
			panic(err)
		}

		// dump to a file
		go func(cs []exchange.CandleStick) {
			jsonResponse, _ := json.Marshal(cs)
			fmt.Fprintf(w, string(jsonResponse))
			err := ioutil.WriteFile("./testbeds/tb1.json", jsonResponse, 0644)
			if err != nil {
				fmt.Println(err)
			}
		}(candleSticks)

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
	
	// fmt.Print(candleSticks)
		

	
	config := configManager.StrategyConfig(strategy)
	config["refresh_frequency"] = "1"
	config["executeAsync"] = "N"
	config["limit_buy"] = "10000"
	config["limit_sell"] = "10000"

	
	client := exchange.NewExchangeProviderFake(&candleSticks, config)
	fmt.Println(client.Balances())

	bot := trading.NewBot(market, strategy, config, client)
	
	uuid := bot.Uuid
	client.OnEnd(func(){
		fmt.Println("STOP")
		message.StopTrader(uuid)
	})
	bot.Start()
	fmt.Println(client.Balances())
	bln,_ := client.Balances()
	jsonResponse, _ := json.Marshal(TestingResult{client.Actions, bln})
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTestbedChart(w http.ResponseWriter, r *http.Request) {
	var candleSticks []exchange.CandleStick
	dat, err := ioutil.ReadFile("./testbeds/tb1.json")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
		panic(err)
	}
	fmt.Print(string(dat))
	err = json.Unmarshal(dat, &candleSticks)
	fmt.Println(err)
	
	var res PlotPoints
	for _, candle := range candleSticks {
		res = append(res, PlotPoint{time.Time(candle.Timestamp).String(), strconv.FormatFloat(candle.Close, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}


func handleTestbedIndicatorChart(w http.ResponseWriter, r *http.Request) {
	indicator := r.URL.Query().Get("name")
	if !utils.StringInSlice(indicator, []string{"ema", "wma", "trima", "rsi", "httrendline"}) {
		panic("indicator is not recognized")
	}
	interval, err := strconv.Atoi(r.URL.Query().Get("interval"))

	var candleSticks []exchange.CandleStick
	dat, err := ioutil.ReadFile("./testbeds/tb1.json")
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
		panic(err)
	}
	fmt.Print(string(dat))
	err = json.Unmarshal(dat, &candleSticks)
	fmt.Println(err)
	
	var closes []float64
	for _, candle := range candleSticks {
		closes = append(closes, candle.Close)
	}	
	
	
	var indicatorData []float64

	fmt.Println("Indicator: ", indicator, interval)
	
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


//Strategies
// Floor finder
// Pump resolver
	