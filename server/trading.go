package main

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math"
	"time"
	"strings"
	"strconv"
	uuid "github.com/satori/go.uuid"
)

type MarketAction struct {
	Action int
	Market string
	Price float64
	Time time.Time
}

const (
	MarketActionBuy = iota
	MarketActionSell = iota
)


type TestingResultAction struct {
	Action int
	Market string
	Price float64
	Time string
}
type TestingResult struct {
	Actions []TestingResultAction
	Balances []PlotPoint
	FinalBalance float64
}

type TradingBot struct {
	market string
	uuid string
	c chan ServerMessage
	tradingConfig map[string]string
	strategy func(string, *[]CandleStick, *MarketAction, map[string]string) *MarketAction
	exchangeProvider ExchangeProvider
	candles []CandleStick
	lastAction *MarketAction
}

func (p TradingBot) Start() {
	defer close(p.c)

	// pre-shot
	p.marketAction()

	ticker := time.NewTicker(30 * time.Second)
L:
	for {
		select {
			case <-ticker.C:
				fmt.Println("Tick", p.market, time.Now().String())
				candleStick, err := p.exchangeProvider.LastCandleStick(p.market, p.tradingConfig["interval"])
				if err != nil {
					fmt.Println("Latest stick was not received: ", err)
				} else {
					fmt.Println("Latest tick", p.candles[len(p.candles)-1], "New tick", candleStick)
					if p.candles[len(p.candles)-1].Timestamp != candleStick.Timestamp {
						fmt.Println("Timestamps different, send the candle")
						temp := append(p.candles, candleStick)
						p.candles = temp[len(temp)-1000:] //take 1000 values
						p.marketAction()
					}
				}
			case msg := <- p.c:
				if msg.Uuid == p.uuid {
					if msg.Action == ServerMessageActionStop {
						fmt.Println("Execution STOP", p.uuid)
						ticker.Stop()
						break L
					}
				}
		}
	}

}

func (p TradingBot) marketAction() {
	marketAction := p.strategy(p.market, &p.candles, p.lastAction, p.tradingConfig)
	if marketAction != nil {
		p.performMarketAction(*marketAction)
	}
}

func (p TradingBot) performMarketAction(action MarketAction) {
	marketSummary, _ := p.exchangeProvider.MarketSummary(action.Market)
	tickers := strings.Split(action.Market, "-")

	if action.Action == MarketActionBuy {		
		amount := amountToOrder(str2flo(p.tradingConfig["limit_buy"]), tickers[0], p.exchangeProvider)
		uuid, _ := p.exchangeProvider.Buy(action.Market, amount, marketSummary.Ask)
		fmt.Println("Order submitted:", uuid, action.Market, marketSummary.Ask)
	} else if action.Action == MarketActionSell {
		amount := amountToOrder(str2flo(p.tradingConfig["limit_sell"]), tickers[1], p.exchangeProvider)
		uuid, _ := p.exchangeProvider.Sell(action.Market, amount, marketSummary.Bid)
		fmt.Println("Order submitted:", uuid, action.Market, marketSummary.Bid)
	} else {
		fmt.Println("Unknown action:", action.Action)
	}
}

func Begin(market string, config map[string]string, strategy func(string, *[]CandleStick, *MarketAction, map[string]string) *MarketAction, exchangeProvider ExchangeProvider) string {

	// periods -> ["oneMin", "fiveMin", "thirtyMin", "hour", "day"]
	candleSticks, err := exchangeProvider.AllCandleSticks(market, config["interval"])
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
		panic(err)
	}
	
	uuid := uuid.NewV4().String()
	traders[uuid] = make(chan ServerMessage)
	bot := TradingBot{market, uuid, traders[uuid], config, strategy, exchangeProvider, candleSticks, nil}

	go bot.Start()
	fmt.Println("Trading started at", time.Now().String()," UUID:", uuid)
	return uuid
}


func strategyWma(market string, candles *[]CandleStick, lastAction *MarketAction, config map[string]string) *MarketAction {
	closes := valuesFromCandles(candles)
	
	wmaMax, err := strconv.Atoi(config["wma_max"])
	handleTradingError(err)
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	minPriceSpike := str2flo(config["min_price_spike"])
	minPriceDip := str2flo(config["min_price_dip"])
	indicatorData1 := talib.Wma(closes, wmaMax)
	indicatorData2 := talib.Wma(closes, wmaMin)

	// if we have a position then we would like to take profits
	if lastAction != nil && lastAction.Action == MarketActionBuy && LastFloat(closes) > LastFloat(indicatorData2) + minPriceSpike {
		return &MarketAction{MarketActionSell, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if LastFloat(closes) < LastFloat(indicatorData2) - minPriceDip {
		return &MarketAction{MarketActionBuy, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}
	
	if ((indicatorData1[len(indicatorData1)-1] < indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] > indicatorData2[len(indicatorData1)-2])) ||	((indicatorData1[len(indicatorData1)-1] > indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] < indicatorData2[len(indicatorData1)-2])) {
		// TODO volume confirmation
		// TODO instrument price above or below wma
		// TODO wait for a retrace

		// TODO sell earlier

		// trend confirmation
		indicatorData3 := talib.HtTrendline(closes)
		if indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1] > 0 { //does it cross above?
			if indicatorData3[len(indicatorData3)-1] > indicatorData3[len(indicatorData3)-2] {
				return &MarketAction{MarketActionBuy, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
			}			
		} else {
			if indicatorData3[len(indicatorData3)-1] < indicatorData3[len(indicatorData3)-2] {
				return &MarketAction{MarketActionSell, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
			}
		}
	} else {
		distance := math.Min(math.Abs(indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1]), math.Abs(indicatorData2[len(indicatorData2)-2] - indicatorData1[len(indicatorData1)-2]))
		fmt.Println("Distance:", distance)
	}
	return nil
}

func strategyDip(market string, candles *[]CandleStick, lastAction *MarketAction, config map[string]string) *MarketAction {
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	minPriceSpike := str2flo(config["min_price_spike"])
	minPriceDip := str2flo(config["min_price_dip"])

	closes := valuesFromCandles(candles)
	// indicatorData1 := talib.Wma(closes, config["wma_max"])
	indicatorData2 := talib.Wma(closes, wmaMin)

	// if we have a position then we would like to take profits
	if lastAction != nil && lastAction.Action == MarketActionBuy && LastFloat(closes) > LastFloat(indicatorData2) + minPriceSpike {
		return &MarketAction{MarketActionSell, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if LastFloat(closes) < LastFloat(indicatorData2) - minPriceDip {
		return &MarketAction{MarketActionBuy, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	fmt.Println("No trading action", LastFloat(closes), LastFloat(indicatorData2))
	return nil
}

func amountToOrder(limit float64, ticker string, client ExchangeProvider) float64 {
	amountToOrder := limit
	balance, err := client.Balance(ticker);
	handleTradingError(err)
	if balance.Available < amountToOrder {
		amountToOrder = balance.Available
	}
	return amountToOrder
}

func valuesFromCandles(candleSticks *[]CandleStick) []float64 {
	var closes []float64
	for _, candle := range *candleSticks {
		closes = append(closes, candle.Close)
	}
	return closes
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func LastFloat(arr []float64) float64 {
	return arr[len(arr) - 1]
}

func str2flo(arg string) float64 {
	r, err := strconv.ParseFloat(arg, 64)
	handleTradingError(err)
	return r
}

func handleTradingError(err error) {
	if err != nil {
		fmt.Println("Trading error: ", err)
		panic(err)
	}
}
//Strategies
// Floor finder
// Pump resolver
	