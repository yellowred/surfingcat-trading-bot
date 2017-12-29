package main

import (
	"log"
	// "fmt"
	"net/http"

	"github.com/rs/cors"
	"github.com/urfave/negroni"
	// configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
	// "github.com/yellowred/surfingcat-trading-bot/server/utils"
	"flag"

	"github.com/gorilla/websocket"
)

var traderStore *message.TraderStore

var (
	apiPort    = flag.String("api-port", "3026", "The API port (i.e. 3026)")
	wssPort    = flag.String("wss-port", "3028", "The WebSocket port (i.e. 3028)")
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	upgraderMt = websocket.TextMessage
)

func main() {
	startServer()
}

func startServer() {

	flag.Parse()
	log.SetFlags(0)

	// start WSS
	wss := http.NewServeMux()
	wss.HandleFunc("/message/", handleWsMessage)
	go func() {
		n := negroni.Classic() // Includes some default middlewares
		n.UseHandler(wss)
		log.Println("Starting to listen on " + *wssPort)
		http.ListenAndServe(":"+*wssPort, n)
	}()

	mux := http.NewServeMux()
	traderStore = message.NewTraderStore()

	mux.HandleFunc("/ema/usdbtc", handleEmaBtcUsd)
	mux.HandleFunc("/chart/usdbtc", handleChartBtcUsd)
	mux.HandleFunc("/indicator", handleIndicatorChart)
	mux.HandleFunc("/trader/start", handleTraderStart)
	mux.HandleFunc("/trader/stop", handleTraderStop)
	mux.HandleFunc("/trader/balance", handleTraderBalance)
	mux.HandleFunc("/trader/status", handleTraderStatus)
	mux.HandleFunc("/strategy/test", handleStrategyTest)
	mux.HandleFunc("/strategy/supertest", handleStrategySuperTest)
	mux.HandleFunc("/chart/testbed", handleTestbedChart)
	mux.HandleFunc("/indicator/testbed", handleTestbedIndicatorChart)

	mux.HandleFunc("/message/", handleMessage)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	n := negroni.Classic() // Includes some default middlewares
	n.Use(c)
	n.UseHandler(mux)

	log.Println("Starting to listen on " + *apiPort)
	log.Fatal(http.ListenAndServe(":"+*apiPort, n))
}
