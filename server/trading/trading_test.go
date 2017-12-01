package trading

import (
	. "github.com/franela/goblin"
	"testing"
	"time"
	"github.com/yellowred/surfingcat-trading-bot/server/exchange"
	message "github.com/yellowred/surfingcat-trading-bot/server/message"
)

type ExchangeProviderFake struct {
	name string
	candles []exchange.CandleStick
}

func (p ExchangeProviderFake) Balances() ([]exchange.Balance, error) {
	balances := []exchange.Balance{exchange.Balance{"BTC", 5, 5, 0, "0x213871928371982", false, "ea56"}}
	return balances, nil
}

func (p ExchangeProviderFake) Balance(ticker string) (exchange.Balance, error) {
	return exchange.Balance{ticker, 5, 5, 0, "0x213871928371982", false, "ea56"}, nil
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
	return p.candles[len(p.candles) - 1], nil
}

func (p ExchangeProviderFake) MarketSummary(market string) (exchange.MarketSummary, error) {
	return exchange.MarketSummary{market, 10000, 6000, 6458, 6450, 6460, 1000, 1000}, nil
}

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

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", 0, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
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

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 1, 1, 100, 100, *time3}}
			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", 0, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": "0.1", "min_price_dip": "0.1"},
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

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 5, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", 0, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": ".1", "min_price_dip": "0.7"},
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

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 20, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionBuy, "USDT-BTC", 10, time.Time(candles[1].Timestamp)},
				map[string]string{"wma_min": "2", "min_price_spike": "0.1", "min_price_dip": "0.1"},
			)
			g.Assert(marketAction.Action).Equal(MarketActionSell)
		})


		g.It("Should not be a sell action if there is no buy action", func() {

			time1 := new(exchange.CandleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(exchange.CandleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(exchange.CandleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 20, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				MarketAction{MarketActionIdle, "", 0, time.Now()},
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
			)
			g.Assert(marketAction.Action).Equal(MarketActionIdle)
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

			candles := []exchange.CandleStick{exchange.CandleStick{20, 10, 10, 1, 100, 100, *time1}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time2}, exchange.CandleStick{20, 10, 10, 1, 100, 100, *time3}}
			
			client := exchange.NewExchangeProviderFake(&candles, map[string]string{"history_size": "1"}, map[string]float64{"USDT": 1000, "BTC": 0})
			traderStore := message.NewTraderStore()
			bot := NewBot(
				"USDT-BTC",
				"dip",
				map[string]string{
					"wma_min": "2", 
					"min_price_spike": "1", 
					"min_price_dip": "1",
					"refresh_frequency": "1",
					"window_size": "3",
				},
				client,
				traderStore,
			)
			client.OnEnd(func(){
				traderStore.Del(bot.Uuid)
			})
			
			bot.Start()
			g.Assert(true).IsTrue()
		})
	})
}
