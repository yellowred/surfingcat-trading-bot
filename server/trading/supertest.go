package trading

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yellowred/surfingcat-trading-bot/server/config"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
	"github.com/yellowred/surfingcat-trading-bot/server/utils"
)

func TestConfigFactory(config map[string]string, variableValues map[string][]string) []map[string]string {
	var updatedConfigs []map[string]string
	for key, valuesArray := range variableValues {
		for _, val := range valuesArray {
			testConfig := utils.CopyMapString(config)
			testConfig[key] = val
			updatedConfigs = append(updatedConfigs, testConfig)
		}
	}
	return updatedConfigs
}

func RunSupertest(candlesData []byte, settingsData []byte, variableParamsData []byte, traderStore *message.TraderStore, logger LoggerInterface) []map[string]string {

	var candleSticks []exchange.CandleStick
	err := json.Unmarshal(candlesData, &candleSticks)
	utils.HandleError(err)

	var settings map[string]string
	err = json.Unmarshal(settingsData, &settings)
	utils.HandleError(err)

	var variableParams map[string][]string
	err = json.Unmarshal(variableParamsData, &variableParams)
	utils.HandleError(err)

	configStrategy := config.StrategyConfig(settings["strategy"])
	configStrategy["refresh_frequency"] = "10000"
	configStrategy["executeAsync"] = "N"
	configStrategy["limit_buy"] = "10000"
	configStrategy["limit_sell"] = "10000"
	// config["window_size"] = "100"

	testConfigs := TestConfigFactory(configStrategy, variableParams)
	/*config["wma_max"]: 50,
	"wma_min": 20,
	"limit_buy": 0.1,
	"limit_sell": 0.1,
	"min_price_spike": 50,
	"min_price_dip": 50*/
	start := time.Now()
	total := 0
	ch := make(chan map[string]string)

	for _, tc := range testConfigs {
		// TODO check params validity
		if ConfigValid(tc) {
			total = total + 1
			go StrategyResult(tc["strategy"], tc["market"], candleSticks, tc, ch, traderStore, logger)
		}
	}
	log.Println(fmt.Sprintf("Launched %d tests.", total))
	var results []map[string]string
	for i := 1; i <= total; i++ {
		item := <-ch
		results = append(results, item)
	}
	sort.Sort(utils.BySuperTestResult(results))

	LogResult(results)

	elapsed := time.Since(start)
	log.Println(fmt.Sprintf("Strategy evaluation took %s\n", elapsed))
	return results
}

func StrategyResult(strategy string, market string, candleSticks []exchange.CandleStick, testConfig map[string]string, ch chan map[string]string, traderStore *message.TraderStore, logger LoggerInterface) {
	tickers := strings.Split(testConfig["market"], "-")
	startBalance := map[string]decimal.Decimal{tickers[0]: utils.Str2dec(testConfig[tickers[0]]), tickers[1]: utils.Str2dec(testConfig[tickers[1]])}

	client := exchange.NewExchangeProviderFake(candleSticks, testConfig, startBalance)

	bot := NewBot(market, strategy, testConfig, &client, traderStore, logger)

	uuid := bot.Uuid
	client.OnEnd(func() {
		traderStore.Del(uuid)
	})

	bot.Start()

	bln, _ := client.Balances()
	logger.PlatformLogger([]string{"finish_bot", uuid, testConfig["wma_max"], testConfig["wma_min"], bln[0].Currency, bln[0].Available.String(), bln[1].Currency, bln[1].Available.String(), candleSticks[len(candleSticks)-1].Close.String()})

	result := bln[0].Available.Div(candleSticks[len(candleSticks)-1].Close).Add(bln[1].Available)
	if bln[0].Currency == "BTC" {
		result = bln[1].Available.Div(candleSticks[len(candleSticks)-1].Close).Add(bln[0].Available)
	}

	testConfig["superTestResult"] = result.String()
	clientActionsJSON, _ := json.Marshal(client.Actions)
	testConfig["clientActions"] = string(clientActionsJSON)
	ch <- testConfig
}

func ConfigValid(config map[string]string) bool {
	if utils.Str2flo(config["wma_max"]) <= utils.Str2flo(config["wma_min"]) {
		return false
	}
	return true
}

func LogResult(results []map[string]string) {
	log.Println("**********************************\nResults:")
	for _, item := range results {
		log.Println(item["wma_max"], item["wma_min"], item["superTestResult"])
	}
}

func CsvResult(results []map[string]string) {
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
}
