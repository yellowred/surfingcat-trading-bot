package trading

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	talib "github.com/markcheno/go-talib"
	uuidGen "github.com/satori/go.uuid"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	"github.com/yellowred/surfingcat-trading-bot/server/message"
	"github.com/yellowred/surfingcat-trading-bot/server/utils"
)

type MarketAction struct {
	Action int
	Market string
	Price  decimal.Decimal
	Time   time.Time
}

const (
	MarketActionBuy  = iota
	MarketActionSell = iota
	MarketActionIdle = iota
)

type TradingBot struct {
	market           string
	Uuid             string
	c                chan message.ServerMessage
	tradingConfig    map[string]string
	strategy         func(string, *[]exchange.CandleStick, MarketAction, map[string]string, func(data []string)) MarketAction
	exchangeProvider exchange.ExchangeProvider
	candles          []exchange.CandleStick
	lastAction       MarketAction
	logger           LoggerInterface
}

type LoggerInterface interface {
	PlatformLogger(message []string)
	BotLogger(botId string, message []string)
	MarketLogger(message []string)
}

func (p *TradingBot) Start() {
	p.logger.PlatformLogger([]string{"start_bot", p.Uuid, "100", "100"}) // temp values 100

	log.Println("Trading started at", time.Now().String(), " UUID:", p.Uuid)
	p.logger.BotLogger(p.Uuid, []string{"start", time.Now().String(), p.market, string(utils.MapStringToJson(p.tradingConfig))})

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
		case msg := <-p.c:
			if msg.Uuid == p.Uuid {
				if msg.Action == message.ServerMessageActionStop {
					p.logger.BotLogger(p.Uuid, []string{"stop", time.Now().String()})
					ticker.Stop()
					break L
				}
			}
		}
	}
}

func (p *TradingBot) marketAction() {
	marketAction := p.strategy(p.market, &p.candles, p.lastAction, p.tradingConfig, func(data []string) {
		// TODO strategy name
		p.logger.BotLogger(p.Uuid, append([]string{"strategy", "dip"}, data...))
	})
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
			uuid, _ := p.exchangeProvider.Buy(p.market, decimal.NewFromFloat(amount).Div(rate), rate)
			p.lastAction = action

			p.logger.BotLogger(p.Uuid, []string{"market_buy", p.market, decimal.NewFromFloat(amount).Div(rate).String(), rate.String()})
			fmt.Println("Order submitted:", p.lastAction, uuid, p.market, amount, rate)
		} else {
			p.logger.BotLogger(p.Uuid, []string{"market_nef", p.market, decimal.NewFromFloat(amount).Div(marketSummary.Ask).String(), marketSummary.Ask.String()})
			fmt.Println("Not enough funds")
		}

	} else if action.Action == MarketActionSell {
		amount := amountToOrder(utils.Str2flo(p.tradingConfig["limit_sell"]), tickers[1], p.exchangeProvider)
		if amount > 0 {
			rate := marketSummary.Bid
			uuid, _ := p.exchangeProvider.Sell(p.market, decimal.NewFromFloat(amount), rate)
			p.lastAction = action

			p.logger.BotLogger(p.Uuid, []string{"market_sell", p.market, utils.Flo2str(amount), rate.String()})
			fmt.Println("Order submitted: SELL", uuid, p.market, amount, rate)
		} else {
			p.logger.BotLogger(p.Uuid, []string{"market_nef", p.market, utils.Flo2str(amount), marketSummary.Ask.String()})
			fmt.Println("Not enough funds")
		}
	} else {
		p.logger.BotLogger(p.Uuid, []string{"market_ua", strconv.Itoa(action.Action)})
		fmt.Println("Unknown action:", action.Action)
	}
}

func NewBot(market, strategy string, config map[string]string, exchangeProvider exchange.ExchangeProvider, traderStore *message.TraderStore, logger LoggerInterface) TradingBot {

	var strategyFunc func(string, *[]exchange.CandleStick, MarketAction, map[string]string, func(data []string)) MarketAction

	switch strategy {
	case "wma":
		strategyFunc = strategyWma
	case "dip":
		strategyFunc = strategyDip
	default:
		panic("Strategy is not recognized")
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

	return TradingBot{market, uuid, ch, testConfig, strategyFunc, exchangeProvider, candleSticks, MarketAction{MarketActionIdle, market, decimal.Zero, time.Now()}, logger}
}

func strategyWma(market string, candles *[]exchange.CandleStick, lastAction MarketAction, config map[string]string, logDecision func(data []string)) MarketAction {
	closes := valuesFromCandles(candles)
	closesFloat := utils.ArrayDecimalToFloat(closes)

	wmaMax, err := strconv.Atoi(config["wma_max"])
	handleTradingError(err)
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	indicatorData1 := talib.Wma(closesFloat, wmaMax)
	indicatorData2 := talib.Wma(closesFloat, wmaMin)

	action := MarketActionIdle
	lastPrice := utils.LastDecimal(closes)
	indi11 := indicatorData1[len(indicatorData1)-1]
	indi12 := indicatorData1[len(indicatorData1)-2]
	indi21 := indicatorData2[len(indicatorData2)-1]
	indi22 := indicatorData2[len(indicatorData2)-2]
	var distance float64 = 0

	if (indi11 < indi21 && indi12 > indi22) || (indi11 > indi21 && indi12 < indi22) {
		// TODO volume confirmation
		// TODO instrument price above or below wma
		// TODO wait for a retrace

		// TODO sell earlier

		// trend confirmation
		indicatorData3 := talib.HtTrendline(closesFloat)
		if indi21-indi11 > 0 { //does it cross above?
			if indicatorData3[len(indicatorData3)-1] > indicatorData3[len(indicatorData3)-2] {
				action = MarketActionBuy
			}
		} else {
			if indicatorData3[len(indicatorData3)-1] < indicatorData3[len(indicatorData3)-2] {
				action = MarketActionSell
			}
		}
	} else {
		distance = math.Min(math.Abs(indi21-indi11), math.Abs(indi22-indi12))
		fmt.Println("Distance:", distance)
	}

	logDecision([]string{
		strconv.Itoa(action),
		strconv.Itoa(lastAction.Action),
		lastPrice.String(),
		utils.Flo2str(indi11),
		utils.Flo2str(indi12),
		utils.Flo2str(indi21),
		utils.Flo2str(indi22),
		utils.Flo2str(distance),
		time.Now().String()})

	return MarketAction{MarketActionIdle, market, utils.LastDecimal(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
}

func strategyDip(market string, candles *[]exchange.CandleStick, lastAction MarketAction, config map[string]string, logDecision func([]string)) MarketAction {
	wmaMin, err := strconv.Atoi(config["wma_min"])
	handleTradingError(err)
	minPriceSpike := utils.Str2flo(config["min_price_spike"])
	minPriceDip := utils.Str2flo(config["min_price_dip"])

	closes := valuesFromCandles(candles)
	closesFloat := utils.ArrayDecimalToFloat(closes)

	// indicatorData1 := talib.Wma(closes, config["wma_max"])
	indicatorData2 := talib.Wma(closesFloat, wmaMin)

	action := MarketActionIdle
	lastPrice := utils.LastDecimal(closes)
	lastIndicator := decimal.NewFromFloat(utils.LastFloat(indicatorData2))

	// fmt.Println(config["wma_max"], config["wma_min"], "Strategy: DIP", lastAction, utils.LastFloat(closes), utils.LastFloat(indicatorData2) + utils.LastFloat(indicatorData2)*minPriceSpike, minPriceDip, minPriceSpike)
	// if we have a position then we would like to take profits
	if lastIndicator.Add(lastIndicator.Mul(decimal.NewFromFloat(minPriceSpike))).LessThan(lastPrice) {
		action = MarketActionSell
	} else
	// if we see some dip we might buy it
	if (lastAction.Action == MarketActionSell || lastAction.Action == MarketActionIdle) && lastPrice.LessThan(lastIndicator.Sub(lastIndicator.Mul(decimal.NewFromFloat(minPriceDip)))) {
		action = MarketActionBuy
	}

	logDecision([]string{
		strconv.Itoa(action),
		strconv.Itoa(lastAction.Action),
		lastPrice.String(),
		lastIndicator.String(),
		utils.Flo2str(minPriceSpike),
		utils.Flo2str(minPriceDip),
		time.Now().String()})
	// fmt.Println("No trading action", utils.LastFloat(closes), utils.LastFloat(indicatorData2), candles)
	return MarketAction{action, market, utils.LastDecimal(closes), time.Time((*candles)[len(*candles)-1].Timestamp)}
}

func amountToOrder(limit float64, ticker string, client exchange.ExchangeProvider) float64 {
	amountToOrder := limit
	balance, err := client.Balance(ticker)
	handleTradingError(err)
	if balance.Available.LessThan(decimal.NewFromFloat(amountToOrder)) {
		amountToOrder, _ = balance.Available.Float64()
	}
	return amountToOrder
}

func valuesFromCandles(candleSticks *[]exchange.CandleStick) []decimal.Decimal {
	var closes []decimal.Decimal
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
