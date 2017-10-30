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

const MIN_PRICE_SPIKE float64 = 50
const MIN_PRICE_DIP float64 = 90

func strategyWma(market string, candles *bittrex.CandleSticks, lastAction *MarketAction) *MarketAction {
	closes := valuesFromCandles(*candles)
	indicatorData1 := talib.Wma(closes, 50)
	indicatorData2 := talib.Wma(closes, 20)

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

func strategyDip(market string, candles *bittrex.CandleSticks, lastAction *MarketAction) *MarketAction {
	closes := valuesFromCandles(*candles)
	// indicatorData1 := talib.Wma(closes, 50)
	indicatorData2 := talib.Wma(closes, 20)

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

func LastFloat(arr []float64) float64 {
	return arr[len(arr) - 1]
}
//Strategies
// Floor finder
// Pump resolver
	