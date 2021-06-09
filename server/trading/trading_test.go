package trading

import (
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/shopspring/decimal"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	message "github.com/yellowred/surfingcat-trading-bot/server/message"
)

type ExchangeProviderFake struct {
	name    string
	candles []exchange.CandleStick
}

func (p ExchangeProviderFake) Balances() ([]exchange.Balance, error) {
	balances := []exchange.Balance{exchange.Balance{"BTC", decimal.New(5, 0), decimal.New(5, 0), decimal.Zero, "0x213871928371982", false, "ea56"}}
	return balances, nil
}

func (p ExchangeProviderFake) Balance(ticker string) (exchange.Balance, error) {
	return exchange.Balance{ticker, decimal.New(5, 0), decimal.New(5, 0), decimal.Zero, "0x213871928371982", false, "ea56"}, nil
}

func (p ExchangeProviderFake) Buy(ticker string, amount float64, rate float64) (string, error) {
	return "yes", nil
}

func (p ExchangeProviderFake) Sell(ticker string, amount float64, rate float64) (string, error) {
	return "yes", nil
}

func (p ExchangeProviderFake) Name() string {
	return exchange.EXCHANGE_PROVIDER_BITTREX
}

func (p ExchangeProviderFake) AllCandleSticks(market string, interval string) ([]exchange.CandleStick, error) {
	return p.candles, nil
}

func (p ExchangeProviderFake) LastCandleStick(market string, interval string) (exchange.CandleStick, error) {
	return p.candles[len(p.candles)-1], nil
}

func (p ExchangeProviderFake) MarketSummary(market string) (exchange.MarketSummary, error) {
	return exchange.MarketSummary{market, decimal.New(10000, 0), decimal.New(6000, 0), decimal.New(6458, 0), decimal.New(6450, 0), decimal.New(6460, 0), decimal.New(1000, 0), decimal.New(1000, 0)}, nil
}

type LoggerFake struct{}

func (p LoggerFake) PlatformLogger(message []string)          {}
func (p LoggerFake) BotLogger(botId string, message []string) {}
func (p LoggerFake) MarketLogger(message []string)            {}

// implement test tables @see https://blog.alexellis.io/golang-writing-unit-tests/
func TestTrading(t *testing.T) {
	g := Goblin(t)

	g.Describe("Trading Strategy DIP", func() {

		g.It("Should be no action on a flat trend", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time1}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time2}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", decimal.Zero, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
				func([]string) {},
			)
			g.Assert(marketAction.Action).Equal(2)
		})

		g.It("Should be a buy action on a dip", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time1}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time2}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time3}}
			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", decimal.Zero, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": "0.1", "min_price_dip": "0.1"},
				func([]string) {},
			)
			g.Assert(marketAction.Action).Equal(MarketActionBuy)
		})

		g.It("Should not be a buy action if the dip is too shallow", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time1}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time2}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(5, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", decimal.Zero, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": ".1", "min_price_dip": "0.7"},
				func([]string) {},
			)
			g.Assert(marketAction.Action).Equal(MarketActionIdle)
		})

		g.It("Should be a sell action on a spike", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time1}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time2}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(20, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionBuy, "USDT-BTC", decimal.New(10, 0), time.Time(candles[1].Timestamp)},
				map[string]string{"wma_min": "2", "min_price_spike": "0.1", "min_price_dip": "0.1"},
				func([]string) {},
			)
			g.Assert(marketAction.Action).Equal(MarketActionSell)
		})
	})

	g.Describe("Trading bot", func() {

		g.It("Should start trading", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time1}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time2}, exchange.CandleStick{decimal.New(20, 0), decimal.New(10, 0), decimal.New(10, 0), decimal.New(1, 0), decimal.New(100, 0), decimal.New(100, 0), *time3}}

			client := exchange.NewExchangeProviderFake(candles, map[string]string{"history_size": "1"}, map[string]decimal.Decimal{"USDT": decimal.New(1000, 0), "BTC": decimal.Zero})
			traderStore := message.NewTraderStore()
			bot := NewBot(
				"USDT-BTC",
				"dip",
				map[string]string{
					"wma_min":           "2",
					"min_price_spike":   "1",
					"min_price_dip":     "1",
					"refresh_frequency": "1",
					"window_size":       "3",
					"limit_sell":        "1000",
				},
				&client,
				traderStore,
				LoggerFake{},
			)
			client.OnEnd(func() {
				traderStore.Del(bot.Uuid)
			})

			bot.Start()
			g.Assert(true).Equal(true)
		})
	})
}
