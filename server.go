package main

import (
	"fmt"
	"net/http"
	"github.com/urfave/negroni"
	"github.com/rs/cors"
)

func main() {
	startServer()
}


func startServer() {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/ema/usdbtc", handleEmaBtcUsd)
	mux.HandleFunc("/chart/usdbtc", handleChartBtcUsd)
	mux.HandleFunc("/indicator", handleIndicatorChart)
	mux.HandleFunc("/trader/start", handleTraderStart)
	mux.HandleFunc("/trader/check", handleTraderCheck)
	mux.HandleFunc("/strategy/test", handleStrategyTest)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	n := negroni.Classic() // Includes some default middlewares
	n.Use(c)
	n.UseHandler(mux)

	fmt.Printf("Starting to listen on %s...\n", ApiPort())
	http.ListenAndServe(":" + ApiPort(), n)
}