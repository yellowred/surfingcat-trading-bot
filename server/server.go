package main

import (
	"fmt"
	"net/http"
	"github.com/urfave/negroni"
	"github.com/rs/cors"
	configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
)

var traderStore *message.TraderStore

func main() {
	startServer()
}


func startServer() {

	mux := http.NewServeMux()
	traderStore = message.NewTraderStore()

	mux.HandleFunc("/ema/usdbtc", handleEmaBtcUsd)
	mux.HandleFunc("/chart/usdbtc", handleChartBtcUsd)
	mux.HandleFunc("/indicator", handleIndicatorChart)
	mux.HandleFunc("/trader/start", handleTraderStart)
	mux.HandleFunc("/trader/stop", handleTraderStop)
	mux.HandleFunc("/trader/check", handleTraderCheck)
	mux.HandleFunc("/trader/balance", handleTraderBalance)
	mux.HandleFunc("/strategy/test", handleStrategyTest)
	mux.HandleFunc("/strategy/supertest", handleStrategySuperTest)
	mux.HandleFunc("/chart/testbed", handleTestbedChart)
	mux.HandleFunc("/indicator/testbed", handleTestbedIndicatorChart)
	
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	n := negroni.Classic() // Includes some default middlewares
	n.Use(c)
	n.UseHandler(mux)




	fmt.Printf("Starting to listen on %s...\n", configManager.ApiPort())
	http.ListenAndServe(":" + configManager.ApiPort(), n)
}