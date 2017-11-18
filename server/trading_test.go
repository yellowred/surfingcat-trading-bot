package main

import (
	. "github.com/franela/goblin"
	"testing"
	"time"
)

type ExchangeProviderFake struct {
	name string
	candles []CandleStick
}

func (p ExchangeProviderFake) Balances() ([]Balance, error) {
	balances := []Balance{Balance{"BTC", 5, 5, 0, "0x213871928371982", false, "ea56"}}
	return balances, nil
}

func (p ExchangeProviderFake) Balance(ticker string) (Balance, error) {
	return Balance{ticker, 5, 5, 0, "0x213871928371982", false, "ea56"}, nil
}

func (p ExchangeProviderFake) Buy(ticker string, amount float64, rate float64) (string, error) {
	return "yes", nil
}

func (p ExchangeProviderFake) Sell(ticker string, amount float64, rate float64) (string, error) {
	return "yes", nil
}

func (p ExchangeProviderFake) Name() string {
	return EXCHANGE_PROVIDER_BITTREX
}

func (p ExchangeProviderFake) AllCandleSticks(market string, interval string) ([]CandleStick, error) {
	return p.candles, nil
}

func (p ExchangeProviderFake) LastCandleStick(market string, interval string) (CandleStick, error) {
	return p.candles[len(p.candles) - 1], nil
}

func (p ExchangeProviderFake) MarketSummary(market string) (MarketSummary, error) {
	return MarketSummary{market, 10000, 6000, 6458, 6450, 6460, 1000, 1000}, nil
}

func TestTrading(t *testing.T) {
	g := Goblin(t)

	g.Describe("Trading Strategy DIP", func() {

		g.It("Should be no action on a flat trend", func() {

			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 10, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				nil,
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
			)
			g.Assert(marketAction == nil).IsTrue()
		})


		g.It("Should be a buy action on a dip", func() {

			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 1, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				nil,
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
			)
			g.Assert(marketAction.Action).Equal(MarketActionBuy)
		})


		g.It("Should not be a buy action if the dip is too shallow", func() {

			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 1, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				nil,
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "3"},
			)
			g.Assert(marketAction == nil).IsTrue()
		})


		g.It("Should be a sell action on a spike", func() {

			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 20, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				&MarketAction{MarketActionBuy, "USDT-BTC", 10, time.Time(candles[1].Timestamp)},
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
			)
			g.Assert(marketAction.Action).Equal(MarketActionSell)
		})


		g.It("Should not be a sell action if there is no buy action", func() {

			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))

			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 20, 1, 100, 100, *time3}}

			marketAction := strategyDip(
				"USDT-BTC",
				&candles,
				nil,
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
			)
			g.Assert(marketAction == nil).IsTrue()
		})
	})

	g.Describe("Trading bot", func() {
		
		g.It("Should start trading", func() {


			time1 := new(candleTime)
			time1.UnmarshalJSON([]byte("2006-01-02T15:04:05"))
			time2 := new(candleTime)
			time2.UnmarshalJSON([]byte("2006-01-03T15:04:05"))
			time3 := new(candleTime)
			time3.UnmarshalJSON([]byte("2006-01-04T15:04:05"))
			candles := []CandleStick{CandleStick{20, 10, 10, 1, 100, 100, *time1}, CandleStick{20, 10, 10, 1, 100, 100, *time2}, CandleStick{20, 10, 10, 1, 100, 100, *time3}}
			
			controlCh := make(chan ServerMessage)
			bot := TradingBot{
				"USDT-BTC",
				"uuid1",
				controlCh,
				map[string]string{"wma_min": "2", "min_price_spike": "1", "min_price_dip": "1"},
				func (market string, candles *[]CandleStick, lastAction *MarketAction, config map[string]string) *MarketAction {
					return nil
				},
				ExchangeProviderFake{"bittrex", candles},
				candles,
				nil,
			}
			g.Assert(bot.uuid).Equal("uuid1")
			go bot.Start()
			// mainContext <- ServerMessage{"uuid1", ServerMessageActionStop}
			g.Assert(true).IsTrue()
		})
	})
		
}
