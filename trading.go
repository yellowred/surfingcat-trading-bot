package main

import (
	"fmt"
	"github.com/thebotguys/golang-bittrex-api/bittrex"
	"github.com/markcheno/go-talib"
	"math"
	"time"
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


type TestingResult struct {
	Actions []MarketAction
	Balances []PlotPoint
	FinalBalance float64
}

func nextTick(market string, candles *bittrex.CandleSticks, tickSource *chan bittrex.CandleStick) {
	candleStick, err := bittrex.GetLatestTick(market, "fiveMin")
	if err != nil {
		fmt.Println("Latest stick was not received: ", err)
	} else {
		temp := *candles
		fmt.Println("Latest tick", temp[len(temp)-1])
		fmt.Println("New tick", candleStick)
		if (temp[len(temp)-1].Timestamp != candleStick.Timestamp) {
			temp = append(temp, *candleStick)
			*candles = temp[len(temp)-1000:] //take 1000 values
			*tickSource <- *candleStick
		}
	}
}


func strategyWma(market string, candles *bittrex.CandleSticks) *MarketAction {
	closes := valuesFromCandles(*candles)
	indicatorData1 := talib.Wma(closes, 100)
	indicatorData2 := talib.Wma(closes, 50)

	if ((indicatorData1[len(indicatorData1)-1] < indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] > indicatorData2[len(indicatorData1)-2])) ||	((indicatorData1[len(indicatorData1)-1] > indicatorData2[len(indicatorData1)-1]) && (indicatorData1[len(indicatorData1)-2] < indicatorData2[len(indicatorData1)-2])) {
		// TODO volume confirmation
		// TODO instrument price above or below wma
		// TODO wait for a retrace
		if indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1] > 0 { //does it cross above?
			return &MarketAction{MarketActionBuy, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
		} else {
			return &MarketAction{MarketActionSell, market, (*candles)[len(*candles)-1].Close, time.Time((*candles)[len(*candles)-1].Timestamp)}
		}
	} else {
		distance := math.Min(math.Abs(indicatorData2[len(indicatorData2)-1] - indicatorData1[len(indicatorData1)-1]), math.Abs(indicatorData2[len(indicatorData2)-2] - indicatorData1[len(indicatorData1)-2]))
		fmt.Println("Distance:", distance)
	}
	return nil
}

func performMarketAction(marketAction MarketAction) {
	if marketAction.Action == MarketActionBuy {
		marketSummary, _ := bittrex.GetMarketSummary(marketAction.Market)
		fmt.Println("WMA crossed, action: BUY", ", price:", marketSummary.Ask)
	} else if marketAction.Action == MarketActionSell {
		marketSummary, _ := bittrex.GetMarketSummary(marketAction.Market)
		fmt.Println("WMA crossed, action: SELL", ", price:", marketSummary.Bid)
	} else {
		fmt.Println("Unknown action:", marketAction.Action)
	}
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
//Strategies
// Floor finder
// Pump resolver
	