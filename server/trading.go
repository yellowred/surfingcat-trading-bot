package main

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math"
	"time"
	"strings"
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
	tradingConfig map[string]string
	strategy func(string, *[]CandleStick, *MarketAction) *MarketAction
	exchangeProvider ExchangeProvider
	candles []CandleStick
	lastAction MarketAction
}

func (p TradingBot) Start() {
	// pre-shot
	p.marketAction()

	tickSource := make(chan CandleStick)
	for {
		select {
			case <-time.After(30 * time.Second):
				fmt.Println("Tick", market, time.Now().String())

				candleStick, err := p.exchangeProvider.LastCandleStick(market, config["interval"])
				if err != nil {
					fmt.Println("Latest stick was not received: ", err)
				} else {
					fmt.Println("Latest tick", temp[len(temp)-1], "New tick", candleStick)
					if (temp[len(temp)-1].Timestamp != candleStick.Timestamp) {
						*tickSource <- *candleStick
					}
				}
			case <-tickSource:
				temp := append(*candles, *candleStick)
				*candles = temp[len(temp)-1000:] //take 1000 values
				p.marketAction()
		}
	}

}

func (p TradingBot) marketAction() {
	marketAction := p.strategy(p.market, &p.candles, &p.lastAction)
	if marketAction != nil {
		p.performMarketAction(*marketAction)
	}
}

func (p TradingBot) performMarketAction(action MarketAction) {
	marketSummary, _ := p.exchangeProvider.MarketSummary(action.Market)
	tickers := strings.Split(action.Market, "-")

	if action.Action == MarketActionBuy {		
		amount := amountToOrder(p.tradingConfig["limit_buy"], tickers[0], p.exchangeProvider)
		uuid, _ := p.exchangeProvider.BuyLimit(action.Market, amount, marketSummary.Ask)
		fmt.Println("Order submitted:", uuid, action.Market, marketSummary.Ask)
	} else if marketAction.Action == MarketActionSell {
		amount := amountToOrder(p.tradingConfig["limit_sell"], tickers[1], p.exchangeProvider)
		uuid, _ := p.exchangeProvider.SellLimit(action.Market, amount, marketSummary.Bid)
		fmt.Println("Order submitted:", uuid, action.Market, marketSummary.Bid)
	} else {
		fmt.Println("Unknown action:", action.Action)
	}
}

func Begin(market string, config map[string]string, strategy func(string, *[]CandleStick, *MarketAction) *MarketAction, exchangeProvider ExchangeProvider) {

	// periods -> ["oneMin", "fiveMin", "thirtyMin", "hour", "day"]
	candleSticks, err := exchangeProvider.GetTicks(market, config["interval"])
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
		panic(err)
	}
	bot := TradingBot{market, config, strategy, exchangeProvider, candleSticks}

	bot.Start()
	fmt.Println("Trading started at", time.Now().String())
}


const MIN_PRICE_SPIKE float64 = 50
const MIN_PRICE_DIP float64 = 90

func strategyWma(market string, candles *[]CandleStick, lastAction *MarketAction, config map[string]string) *MarketAction {
	closes := valuesFromCandles(*candles)
	indicatorData1 := talib.Wma(closes, config["wma_max"])
	indicatorData2 := talib.Wma(closes, config["wma_min"])

	// if we have a position then we would like to take profits
	if lastAction != nil && lastAction.Action == MarketActionBuy && LastFloat(closes) > LastFloat(indicatorData2) + MIN_PRICE_SPIKE {
		return &MarketAction{MarketActionSell, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if LastFloat(closes) < LastFloat(indicatorData2) - MIN_PRICE_DIP {
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
	closes := valuesFromCandles(*candles)
	// indicatorData1 := talib.Wma(closes, config["wma_max"])
	indicatorData2 := talib.Wma(closes, config["wma_min"])

	// if we have a position then we would like to take profits
	if lastAction != nil && lastAction.Action == MarketActionBuy && LastFloat(closes) > LastFloat(indicatorData2) + MIN_PRICE_SPIKE {
		return &MarketAction{MarketActionSell, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if LastFloat(closes) < LastFloat(indicatorData2) - MIN_PRICE_DIP {
		return &MarketAction{MarketActionBuy, market, LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	return nil
}

func amountToOrder(limit float64, ticker string, client ExchangeProvider) float64 {
	amountToOrder := config["limit_sell"]
	balance, err := client.Balance(tickers[1]);
	handleTradingError(err)
	if balance.Available < amountToOrder {
		amountToOrder = balance.Available
	}
	return amountToOrder
}

func valuesFromCandles(candleSticks bittrex.CandleSticks) []float64 {
	var closes []float64
	for _, candle := range candleSticks {
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


func handleTradingError(err error) {
	if err != nil {
		fmt.Println("Trading error: ", err)
		panic(err)
	}
}
//Strategies
// Floor finder
// Pump resolver
	