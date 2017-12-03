package trading

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math"
	"time"
	"strings"
	"strconv"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
	uuidGen "github.com/satori/go.uuid"
	"github.com/yellowred/surfingcat-trading-bot/server/utils"
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
	MarketActionIdle = iota
)


type TradingBot struct {
	market string
	Uuid string
	c chan message.ServerMessage
	tradingConfig map[string]string
	strategy func(string, *[]exchange.CandleStick, MarketAction, map[string]string) MarketAction
	exchangeProvider exchange.ExchangeProvider
	candles []exchange.CandleStick
	lastAction MarketAction
}

func (p *TradingBot) Start() {
	
	fmt.Println("Trading started at", time.Now().String()," UUID:", p.Uuid)

	// pre-shot
	p.marketAction()

	frequency, err := strconv.Atoi(p.tradingConfig["refresh_frequency"])
	handleTradingError(err)
	windowSize, err := strconv.Atoi(p.tradingConfig["window_size"])
	handleTradingError(err)
	ticker := time.NewTicker(time.Duration(frequency) * time.Microsecond)
L:
	for {
		select {
			case <-ticker.C:
				// fmt.Println("Tick", p.market, time.Now().String())
				candleStick, err := p.exchangeProvider.LastCandleStick(p.market, p.tradingConfig["interval"])
				if err != nil {
					fmt.Println("Latest stick was not received: ", err)
				} else {
					// fmt.Println("***********************************************")
					// fmt.Println("Latest tick", p.candles[len(p.candles)-1], "New tick", candleStick)
					if p.candles[len(p.candles)-1].Timestamp != candleStick.Timestamp {
						// fmt.Println("Timestamps different, send the candle")
						temp := append(p.candles, candleStick)
						// fmt.Println(len(temp), windowSize)
						p.candles = temp[len(temp)-windowSize:] //take windowSize values
						p.marketAction()
					}
				}
			case msg := <- p.c:
				if msg.Uuid == p.Uuid {
					if msg.Action == message.ServerMessageActionStop {
						ticker.Stop()
						break L
					}
				}
		}
	}
}

func (p *TradingBot) marketAction() {
	marketAction := p.strategy(p.market, &p.candles, p.lastAction, p.tradingConfig)
	if marketAction.Action != MarketActionIdle {
		p.performMarketAction(marketAction)
	}
}

func (p *TradingBot) performMarketAction(action MarketAction) {
	marketSummary, _ := p.exchangeProvider.MarketSummary(p.market)
	tickers := strings.Split(p.market, "-")

	//	"Bid":7666.01000001,"Ask":7675.75100000
	// TODO: trade is submitted but market can move and the order will be not executed
	if action.Action == MarketActionBuy {		
		amount := amountToOrder(utils.Str2flo(p.tradingConfig["limit_buy"]), tickers[0], p.exchangeProvider)
		if amount > 0 {
			rate := marketSummary.Ask
			uuid, _ := p.exchangeProvider.Buy(p.market, amount/rate, rate)
			p.lastAction = action
			fmt.Println("Order submitted:", p.lastAction, uuid, p.market, amount, rate)
		} else {
			fmt.Println("Not enough funds")
		}
		
	} else if action.Action == MarketActionSell {
		amount := amountToOrder(utils.Str2flo(p.tradingConfig["limit_sell"]), tickers[1], p.exchangeProvider)
		if amount > 0 {
			rate := marketSummary.Bid
			uuid, _ := p.exchangeProvider.Sell(p.market, amount, rate)
			p.lastAction = action
			fmt.Println("Order submitted: SELL", uuid, p.market, amount, rate)
		} else {
			fmt.Println("Not enough funds")
		}
	} else {
		fmt.Println("Unknown action:", action.Action)
	}
}


func NewBot(market string, strategy string, config map[string]string, exchangeProvider exchange.ExchangeProvider, traderStore *message.TraderStore) TradingBot {

	var strategyFunc func(string, *[]exchange.CandleStick, MarketAction, map[string]string) MarketAction
	switch strategy {
	case "wma": strategyFunc = strategyWma
	case "dip": strategyFunc = strategyDip
	default: panic("Strategy is not recognized")
	}
	// periods -> ["oneMin", "fiveMin", "thirtyMin", "hour", "day"]
	testConfig := utils.CopyMapString(config)
	candleSticks, err := exchangeProvider.AllCandleSticks(market, testConfig["interval"])
	if err != nil {
		fmt.Println("ERROR OCCURRED: ", err)
		panic(err)
	}
	uuid := uuidGen.NewV4().String()
	ch := traderStore.Add(uuid)

	
	return TradingBot{market, uuid, ch, testConfig, strategyFunc, exchangeProvider, candleSticks, MarketAction{MarketActionIdle, market, 0, time.Now()}}
}


func strategyWma(market string, candles *[]exchange.CandleStick, lastAction MarketAction, config map[string]string) MarketAction {
	closes := valuesFromCandles(candles)
	
	wmaMax, err := strconv.Atoi(config["wma_max"])
	handleTradingError(err)
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	minPriceSpike := utils.Str2flo(config["min_price_spike"])
	minPriceDip := utils.Str2flo(config["min_price_dip"])
	indicatorData1 := talib.Wma(closes, wmaMax)
	indicatorData2 := talib.Wma(closes, wmaMin)

	// if we have a position then we would like to take profits
	if lastAction.Action == MarketActionBuy && utils.LastFloat(closes) > utils.LastFloat(indicatorData2) + minPriceSpike {
		return MarketAction{MarketActionSell, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if utils.LastFloat(closes) < utils.LastFloat(indicatorData2) - minPriceDip {
		return MarketAction{MarketActionBuy, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
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
				return MarketAction{MarketActionBuy, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
			}			
		} else {
			if indicatorData3[len(indicatorData3)-1] < indicatorData3[len(indicatorData3)-2] {
				return MarketAction{MarketActionSell, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
			}
		}
	} else {
		distance := math.Min(math.Abs(indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1]), math.Abs(indicatorData2[len(indicatorData2)-2] - indicatorData1[len(indicatorData1)-2]))
		fmt.Println("Distance:", distance)
	}
	return MarketAction{MarketActionIdle, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
}

func strategyDip(market string, candles *[]exchange.CandleStick, lastAction MarketAction, config map[string]string) MarketAction {
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	minPriceSpike := utils.Str2flo(config["min_price_spike"])
	minPriceDip := utils.Str2flo(config["min_price_dip"])

	closes := valuesFromCandles(candles)
	// indicatorData1 := talib.Wma(closes, config["wma_max"])
	indicatorData2 := talib.Wma(closes, wmaMin)

	// fmt.Println(config["wma_max"], config["wma_min"], "Strategy: DIP", lastAction, utils.LastFloat(closes), utils.LastFloat(indicatorData2) + utils.LastFloat(indicatorData2)*minPriceSpike, minPriceDip, minPriceSpike)
	// if we have a position then we would like to take profits
	if (lastAction.Action == MarketActionBuy) && utils.LastFloat(closes) > utils.LastFloat(indicatorData2) + utils.LastFloat(indicatorData2)*minPriceSpike {
		return MarketAction{MarketActionSell, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// if we see some dip we might buy it
	if (lastAction.Action == MarketActionSell || lastAction.Action == MarketActionIdle) && utils.LastFloat(closes) < utils.LastFloat(indicatorData2) - utils.LastFloat(indicatorData2)*minPriceDip {
		return MarketAction{MarketActionBuy, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
	}

	// fmt.Println("No trading action", utils.LastFloat(closes), utils.LastFloat(indicatorData2), candles)
	return MarketAction{MarketActionIdle, market, utils.LastFloat(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
}

func amountToOrder(limit float64, ticker string, client exchange.ExchangeProvider) float64 {
	amountToOrder := limit
	balance, err := client.Balance(ticker);
	handleTradingError(err)
	if balance.Available < amountToOrder {
		amountToOrder = balance.Available
	}
	return amountToOrder
}

func valuesFromCandles(candleSticks *[]exchange.CandleStick) []float64 {
	var closes []float64
	for _, candle := range *candleSticks {
		closes = append(closes, candle.Close)
	}
	return closes
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
	