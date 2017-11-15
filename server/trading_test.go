package main

import (
	. "github.com/franela/goblin"
	"testing"
)

func TestTrading(t *testing.T) {
	g := Goblin(t)
	g.Describe("Trading Strategy DIP", func() {
		// Passing Test
		g.It("Should be no action on flat trend", func() {

			time1, err := time.Parse("2016-11-02", "2016-11-02")
			time2, err := time.Parse("2016-11-02", "2016-11-02")
			time3, err := time.Parse("2016-11-02", "2016-11-02")

			marketAction := strategyDip(
				"USDT-BTC",
				[3]CandleStick{CandleStick{20, 10, 10, 1, 100, 100, time1}, CandleStick{20, 10, 10, 1, 100, 100, time2}, CandleStick{20, 10, 10, 1, 100, 100, time3}},
				nil,
				map[string]string{"wma_min": 50},
			)
			g.Assert(marketAction).Equal(nil)
		})
	})
}
