package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
)

var (
	candlesSource = flag.String("provider", "bittrex", "Source of candles (bittrex).")
	market        = flag.String("market", "USDT-BTC", "Market string (USDT-BTC, BTC-LTC, etc).")
)

// How to use: `echo to_be_encrypted | go run tools/passwords.go
func main() {
	flag.Parse()

	provider := ""
	if *candlesSource == "bittrex" {
		provider = exchange.EXCHANGE_PROVIDER_BITTREX
	} else {
		log.Fatalln("Unknown candles source.")
	}
	btx := exchange.ExchangeClient(provider, configManager.ExchangeConfig(provider))

	// get data
	var candleSticks []exchange.CandleStick
	var err error
	candleSticks, err = btx.AllCandleSticks(*market, "fiveMin")
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	jsonResponse, _ := json.Marshal(candleSticks)
	fmt.Println(string(jsonResponse))
}
