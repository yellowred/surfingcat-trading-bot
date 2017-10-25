package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/urfave/negroni"
	"github.com/rs/cors"
)

func main() {
	startServer(env("API_PORT", "3026"))
}


func startServer(port string) {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/ema/usdbtc", handleEmaBtcUsd)
	mux.HandleFunc("/chart/usdbtc", handleChartBtcUsd)
	mux.HandleFunc("/indicator", handleIndicatorChart)
	mux.HandleFunc("/trader/start", handleTraderStart)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	n := negroni.Classic() // Includes some default middlewares
	n.Use(c)
	n.UseHandler(mux)

	fmt.Printf("Starting to listen on %s...\n", port)
	http.ListenAndServe(":"+port, n)
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}