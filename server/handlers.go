package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	jwt "github.com/dgrijalva/jwt-go"
	talib "github.com/markcheno/go-talib"
	"github.com/yellowred/golang-bittrex-api/bittrex"
	configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	trading "github.com/yellowred/surfingcat-trading-bot/server/trading"
	"github.com/yellowred/surfingcat-trading-bot/server/utils"
	"golang.org/x/crypto/bcrypt"
)

type PlotPoint struct {
	Date  string
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
	if err != nil || interval < 5 {
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
	if err != nil || interval < 5 {
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
	case "ema":
		indicatorData = talib.Ema(closes, interval)
	case "wma":
		indicatorData = talib.Wma(closes, interval)
	case "trima":
		indicatorData = talib.Trima(closes, interval)
	case "rsi":
		indicatorData = talib.Rsi(closes, interval)
	case "httrendline":
		indicatorData = talib.HtTrendline(closes)
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

	bot := trading.NewBot(market, strategy, configManager.StrategyConfig(strategy), exchange.ExchangeClient(exchange.EXCHANGE_PROVIDER_BITTREX, configManager.ExchangeConfig(exchange.EXCHANGE_PROVIDER_BITTREX)), traderStore)
	go bot.Start()
	jsonResponse, _ := json.Marshal(bot.Uuid)
	fmt.Fprintf(w, string(jsonResponse))
}

func handleTraderStop(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	traderStore.Del(uuid)
	jsonResponse, _ := json.Marshal(uuid)
	fmt.Fprintf(w, string(jsonResponse))
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

func handleTraderStatus(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(utils.Bots())
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
	Actions  []exchange.TestMarketAction
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
			err := ioutil.WriteFile("./testbeds/tb2.json", jsonResponse, 0644)
			if err != nil {
				fmt.Println(err)
			}
		}(candleSticks)

	} else {
		dat := configManager.TestbedFile(market)
		err := json.Unmarshal(dat, &candleSticks)
		utils.HandleError(err)
		market = configManager.TestbedMarket(market)
	}

	// fmt.Print(candleSticks)

	config := configManager.StrategyConfig(strategy)
	config["refresh_frequency"] = "1"
	config["executeAsync"] = "N"
	config["limit_buy"] = "10000"
	config["limit_sell"] = "10000"

	start := time.Now()
	ch := make(chan map[string]string)
	testConfig := utils.CopyMapString(config)
	testConfig["wma_max"] = "20"
	testConfig["wma_min"] = "2"
	go StrategyResult(strategy, market, candleSticks, testConfig, ch)

	var results []map[string]string
	item := <-ch
	results = append(results, item)

	fmt.Println("**********************************\nResults:")
	for _, item := range results {
		fmt.Println(item["wma_max"], item["wma_min"], item["superTestResult"])
	}

	elapsed := time.Since(start)
	fmt.Printf("Strategy evaluation took %s\n", elapsed)

	jsonResponse, _ := json.Marshal(results)
	fmt.Fprintf(w, string(jsonResponse))
}

func handleTestbedChart(w http.ResponseWriter, r *http.Request) {
	var candleSticks []exchange.CandleStick

	market := r.URL.Query().Get("market")
	dat := configManager.TestbedFile(market)
	err := json.Unmarshal(dat, &candleSticks)
	utils.HandleError(err)

	var res PlotPoints
	for _, candle := range candleSticks {
		c, _ := candle.Close.Float64()
		res = append(res, PlotPoint{time.Time(candle.Timestamp).String(), strconv.FormatFloat(c, 'f', 6, 64)})
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
	market := r.URL.Query().Get("market")
	dat := configManager.TestbedFile(market)
	err = json.Unmarshal(dat, &candleSticks)
	utils.HandleError(err)

	var closes []float64
	for _, candle := range candleSticks {
		c, _ := candle.Close.Float64()
		closes = append(closes, c)
	}

	var indicatorData []float64

	fmt.Println("Indicator: ", indicator, interval)

	switch indicator {
	case "ema":
		indicatorData = talib.Ema(closes, interval)
	case "wma":
		indicatorData = talib.Wma(closes, interval)
	case "trima":
		indicatorData = talib.Trima(closes, interval)
	case "rsi":
		indicatorData = talib.Rsi(closes, interval)
	case "httrendline":
		indicatorData = talib.HtTrendline(closes)
	}

	var res PlotPoints
	for i, indicatorValue := range indicatorData {
		res = append(res, PlotPoint{time.Time(candleSticks[i].Timestamp).String(), strconv.FormatFloat(indicatorValue, 'f', 6, 64)})
	}

	jsonResponse, _ := json.Marshal(res)
	fmt.Fprintf(w, string(jsonResponse))
}

func handleStrategySuperTest(w http.ResponseWriter, r *http.Request) {

	strategy := r.URL.Query().Get("strategy")

	// get data
	var candleSticks []exchange.CandleStick

	market := r.URL.Query().Get("market")
	dat := configManager.TestbedFile(market)
	err := json.Unmarshal(dat, &candleSticks)
	utils.HandleError(err)
	market = configManager.TestbedMarket(market)

	config := configManager.StrategyConfig(strategy)
	config["refresh_frequency"] = "10000"
	config["executeAsync"] = "N"
	config["limit_buy"] = "10000"
	config["limit_sell"] = "10000"
	// config["window_size"] = "100"

	/*config["wma_max"]: 50,
	"wma_min": 20,
	"limit_buy": 0.1,
	"limit_sell": 0.1,
	"min_price_spike": 50,
	"min_price_dip": 50*/
	start := time.Now()
	total := 0
	ch := make(chan map[string]string)
	for _, wmaMax := range append(utils.ARange(10, 20, 2), utils.ARange(20, 100, 10)...) {
		// for _, wmaMax := range utils.ARange(80, 100, 10) {
		for _, wmaMin := range append(utils.ARange(2, 10, 1), utils.ARange(10, 50, 5)...) {
			// for _, wmaMin := range utils.ARange(2, 4, 2) {

			testConfig := utils.CopyMapString(config)
			testConfig["wma_max"] = strconv.FormatInt(wmaMax, 10)
			testConfig["wma_min"] = strconv.FormatInt(wmaMin, 10)
			if wmaMax > wmaMin {
				fmt.Println("ITERATE", wmaMax, wmaMin, testConfig, total)
				total = total + 1
				go StrategyResult(strategy, market, candleSticks, testConfig, ch)
			}

		}
	}

	// 47c89d79-3c47-42f2-a781-59a836c3df0d

	var results []map[string]string

	for i := 1; i <= total; i++ {
		item := <-ch
		results = append(results, item)
	}

	sort.Sort(utils.BySuperTestResult(results))

	fmt.Println("**********************************\nResults:")
	for _, item := range results {
		fmt.Println(item["wma_max"], item["wma_min"], item["superTestResult"])
	}

	matrix := make(map[string]map[string]string)
	for _, item := range results {
		// fmt.Println(item["superTestResult"], item["wma_max"], item["wma_min"])
		if matrix[item["wma_max"]] == nil {
			matrix[item["wma_max"]] = make(map[string]string)
		}
		matrix[item["wma_max"]][item["wma_min"]] = item["superTestResult"]
	}

	csv := ""
	for _, wmaMax := range append(utils.ARange(10, 20, 2), utils.ARange(20, 100, 10)...) {
		row := strconv.FormatInt(wmaMax, 10) + ","
		for _, wmaMin := range append(utils.ARange(2, 10, 1), utils.ARange(10, 50, 5)...) {
			wmaMaxS := strconv.FormatInt(wmaMax, 10)
			wmaMinS := strconv.FormatInt(wmaMin, 10)
			row = row + "," + matrix[wmaMaxS][wmaMinS]
		}
		csv = csv + row + "\n"
	}

	fmt.Println("**********************************\nCSV:")
	fmt.Println(csv)

	elapsed := time.Since(start)
	fmt.Printf("Strategy evaluation took %s\n", elapsed)

	jsonResponse, _ := json.Marshal(matrix)
	fmt.Fprintf(w, string(jsonResponse))
}

func StrategyResult(strategy string, market string, candleSticks []exchange.CandleStick, conf map[string]string, ch chan map[string]string) {
	tickers := strings.Split(market, "-")
	startBalance := map[string]decimal.Decimal{tickers[0]: decimal.Zero, tickers[1]: decimal.Zero}
	startBalance["BTC"] = decimal.New(1, 0)
	client := exchange.NewExchangeProviderFake(candleSticks, conf, startBalance)

	bot := trading.NewBot(market, strategy, conf, &client, traderStore)
	uuid := bot.Uuid
	client.OnEnd(func() {
		traderStore.Del(uuid)
	})

	utils.Logger.PlatformLogger([]string{"start_bot", uuid, conf["wma_max"], conf["wma_min"]})

	fmt.Println("****************************\nSTART BOT", bot.Uuid, conf["wma_max"], conf["wma_min"], "****************************")
	bot.Start()

	bln, _ := client.Balances()
	jsonResponse, _ := json.Marshal(client.Actions)
	utils.Logger.PlatformLogger([]string{"finish_bot", uuid, conf["wma_max"], conf["wma_min"], bln[0].Currency, bln[0].Available.String(), bln[1].Currency, bln[1].Available.String(), candleSticks[len(candleSticks)-1].Close.String()})
	fmt.Println("****************************\nFINISH BOT", bot.Uuid, conf["wma_max"], conf["wma_min"], bln, candleSticks[len(candleSticks)-1].Close, string(jsonResponse), "****************************")

	result := bln[0].Available.Div(candleSticks[len(candleSticks)-1].Close).Add(bln[1].Available)
	if bln[0].Currency == "BTC" {
		result = bln[1].Available.Div(candleSticks[len(candleSticks)-1].Close).Add(bln[0].Available)
	}

	conf["superTestResult"] = result.String()
	ch <- conf
}

//Strategies
// Floor finder
// Pump resolver

func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/message/platform" {
		r := utils.ConsumeMessages()
		jsonResponse, _ := json.Marshal(r)
		fmt.Fprintf(w, string(jsonResponse))
	} else {
		http.NotFound(w, r)
		return
	}
}

func handleWsMessage(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	messages := utils.ConsumeMessages()
	for {
		select {
		case msg := <-messages:
			jsonResponse, _ := json.Marshal(map[string]string{
				"topic": msg.Topic,
				"value": string(msg.Value),
			})
			err = c.WriteMessage(upgraderMt, jsonResponse)
			if err != nil {
				log.Println("ws_write:", err)
				break
			}
		}
	}
}

// UserSignup -
func handleUserSignup(w http.ResponseWriter, req *http.Request) {
	/*
		decoder := json.NewDecoder(req.Body)
		jsondata := UserJSON{}
		err := decoder.Decode(&jsondata)

		if err != nil || jsondata.Username == "" || jsondata.Password == "" {
			http.Error(w, "Missing username or password", http.StatusBadRequest)
			return
		}

		if api.users.HasUser(jsondata.Username) {
			http.Error(w, "username already exists", http.StatusBadRequest)
			return
		}

		user := api.users.AddUser(jsondata.Username, jsondata.Password)

		jsontoken := auth.GetJSONToken(user)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsontoken))
	*/
}

func handleUserLogin(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	jsondata := utils.User{}
	err := decoder.Decode(&jsondata)

	log.Println(err, jsondata)
	if err != nil || jsondata.Login == "" || jsondata.Password == "" {
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}

	/*
		TODO sign-up code
		hash, err := bcrypt.GenerateFromPassword([]byte("matsumura2121"), bcrypt.DefaultCost)
		if err != nil {
			panic("Permissions: bcrypt password hashing unsuccessful")
		}
		log.Println(string(hash))
	*/

	user := utils.FindUser(jsondata.Login)
	if user.Login == "" {
		http.Error(w, "login not found", http.StatusBadRequest)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(jsondata.Password)) != nil {
		http.Error(w, "bad password", http.StatusBadRequest)
		return
	} else {

	}

	jsontoken := GetJSONToken(user)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsontoken))

}

var signingKey = []byte("x-sign-key")

// GetToken create a jwt token with user claims
func GetToken(user utils.User) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uuid"] = user.Uuid
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	signedToken, _ := token.SignedString(signingKey)
	return signedToken
}

// GetJSONToken create a JSON token string
func GetJSONToken(user utils.User) string {
	token := GetToken(user)
	jsontoken := "{\"id_token\": \"" + token + "\"}"
	return jsontoken
}
