package main

import (
	"log"
	"os"
	// "fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	mgo "gopkg.in/mgo.v2"
	// configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/kafka"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
	"github.com/yellowred/surfingcat-trading-bot/server/mongo"
	// "github.com/yellowred/surfingcat-trading-bot/server/utils"
	"flag"

	gmux "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var traderStore *message.TraderStore
var kafkaLogger *kafka.KafkaLogger
var stateStorage *mongo.MongoStateStorage

var (
	apiPort = flag.String("api-port", "3026", "The API port (i.e. 3026)")

	wssPort    = flag.String("wss-port", "3028", "The WebSocket port (i.e. 3028)")
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	upgraderMt = websocket.TextMessage
)

var (
	kafkaConn     = flag.String("kafka-host", "", "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181,zookeeper3.local:2181`)")
	consumerGroup = flag.String("kafka-consumer-group", "group.testing", "The name of the consumer group, used for coordination and load balancing")
	zookeeper     = flag.String("kafka-zookeeper-host", "", "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181,zookeeper3.local:2181`)")

	zookeeperNodes []string
)

var (
	mongoConn    = flag.String("mongo-host", "", "A comma-separated MongoDB connection string (e.g. `192.168.10.100:27017`).")
	mongoDebug   = flag.String("mongo-debug", "false", "Debug mongo: true/false.")
	sessionMongo *mgo.Session
)

func main() {
	flag.Parse()

	traderStore = message.NewTraderStore()

	kafkaLogger = kafka.NewLogger(*kafkaConn)

	stateStorage = mongo.NewMongoStateStorage(*mongoConn, *mongoDebug == "true")

	startServer()
}

func startServer() {

	log.SetFlags(0)

	// go StartWss()

	router := gmux.NewRouter()
	setupRoutes(router)

	n := negroni.Classic()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"X-Requested-With", "authorization", "Content-Type"},
		Debug:            true,
		AllowCredentials: true,
	})
	c.Log = log.New(os.Stdout, "[Cors] ", log.LstdFlags)

	n.Use(c)
	n.UseHandler(router)
	n.Run(":" + *apiPort)

	/*
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
	*/
}

/*
func StartWss() {
	wss := http.NewServeMux()
	wss.HandleFunc("/message/", handleWsMessage)
	n := negroni.Classic() // Includes some default middlewares
	n.UseHandler(wss)
	log.Println("Starting to listen on " + *wssPort)
	http.ListenAndServe(":"+*wssPort, n)
}
*/

func setupRoutes(r *gmux.Router) {
	a := r.PathPrefix("/api").Subrouter()
	u := a.PathPrefix("/user").Subrouter()
	u.HandleFunc("/signup", handleUserSignup).Methods("POST")
	u.HandleFunc("/login", handleUserLogin).Methods("POST")

	mw := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("x-sign-key"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	a.Handle("/trader/status", negroni.New(
		negroni.HandlerFunc(mw.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(handleTraderStatus)),
	))

	a.HandleFunc("/server/status", handleServerStatus).Methods("GET")

	// r.HandleFunc("/strategy/test", handleStrategyTest).Methods("GET")
	r.HandleFunc("/strategy/supertest", handleStrategySuperTest).Methods("GET")
	// r.HandleFunc("/indicator", handleIndicatorChart).Methods("GET")
}
