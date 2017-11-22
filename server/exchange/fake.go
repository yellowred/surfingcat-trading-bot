package exchange

import (
	"strings"
	"errors"
	"fmt"
	"time"
)

const HISTORY_SIZE = 1000
const EXCHANGE_COMISSION = 0.0025


type TestMarketAction struct {
	Action int
	Value float64
	Date string
}


type ExchangeProviderFake struct {
	testbed *[]CandleStick
	config map[string]string
	candles []CandleStick
	index int
	balances map[string]float64
	onEndCallback func()
	Actions []TestMarketAction
}

func (p *ExchangeProviderFake) Balances() ([]Balance, error) {
	var balances []Balance
	for currency, bln := range p.balances  {
		balances = append(balances, Balance{currency, bln, bln, bln, "0x", true, "uuid"+currency})
	}
	return balances, nil
}

func (p *ExchangeProviderFake) Balance(ticker string) (Balance, error) {
	
	if bln, ok := p.balances[ticker]; ok {
		return Balance{ticker, bln, bln, 0, "0x", true, "uuid"+ticker}, nil
	}
	return Balance{}, nil
}

func (p *ExchangeProviderFake) Buy(market string, amount float64, rate float64) (string, error) {
	tickers := strings.Split(market, "-")

	var (
		uuid string
		err error
	)

	fmt.Println("SELL_B", p.balances, tickers, amount, rate)
	// if there is liquidity
	if p.balances[tickers[0]] >= amount * rate {
		// add to balance
		p.balances[tickers[1]] = p.balances[tickers[1]] + amount - amount * EXCHANGE_COMISSION 
		p.balances[tickers[0]] = p.balances[tickers[0]] - amount * rate
		
		candle := (*p.testbed)[HISTORY_SIZE + p.index]
		p.Actions = append(p.Actions, TestMarketAction{0, rate, time.Time(candle.Timestamp).String()})
		
		uuid = "OK_" + market; err = nil 
	} else {
		fmt.Println("INSUFFICIENT_FUNDS", p.balances, amount, rate)
		uuid = "FAIL_" + market; err = errors.New("INSUFFICIENT_FUNDS") 
	}
	fmt.Println("Balance", p.balances)
	return uuid, err
}

func (p *ExchangeProviderFake) Sell(market string, amount float64, rate float64) (string, error) {
	tickers := strings.Split(market, "-")
	
	var (
		uuid string
		err error
	)

	fmt.Println("SELL_A", p.balances, tickers, amount, rate)
	// if there is liquidity
	if p.balances[tickers[1]] >= amount {
		// add to balance
		p.balances[tickers[1]] = p.balances[tickers[1]] - amount
		p.balances[tickers[0]] = p.balances[tickers[0]] + amount * rate  - amount * rate * EXCHANGE_COMISSION
		
		candle := (*p.testbed)[HISTORY_SIZE + p.index]
		p.Actions = append(p.Actions, TestMarketAction{1, rate, time.Time(candle.Timestamp).String()})

		uuid = "OK_" + market; err = nil
	} else {
		fmt.Println("INSUFFICIENT_FUNDS", p.balances, amount, rate)
		uuid = "FAIL_" + market; err = errors.New("INSUFFICIENT_FUNDS") 
	}
	fmt.Println("Balance", p.balances)
	return uuid, err
}

func (p *ExchangeProviderFake) Name() string {
	return EXCHANGE_PROVIDER_FAKE
}

func (p *ExchangeProviderFake) AllCandleSticks(market string, interval string) ([]CandleStick, error) {
	return p.candles, nil
}

func (p *ExchangeProviderFake) LastCandleStick(market string, interval string) (CandleStick, error) {
	if p.index<len(*(p.testbed)) - HISTORY_SIZE - 1 {
		p.index++
		fmt.Println("Candle #", p.index)
	} else {
		if p.onEndCallback != nil {
			go p.onEndCallback()
		}
	}
	return (*p.testbed)[HISTORY_SIZE + p.index], nil
}

func (p *ExchangeProviderFake) MarketSummary(market string) (MarketSummary, error) {

	// CandleStick{rBittrex.High, rBittrex.Open, rBittrex.Close, rBittrex.Low, rBittrex.Volume, rBittrex.BaseVolume, t}
	candle, _ := p.LastCandleStick(market, "")
	return MarketSummary{market, candle.High, candle.Low, candle.Close, candle.Close*0.999, candle.Close*1.001, candle.Volume, candle.BaseVolume}, nil
}

func (p *ExchangeProviderFake) OnEnd(cb func()) {
	p.onEndCallback = cb
}

func NewExchangeProviderFake(testbed *[]CandleStick, config map[string]string) *ExchangeProviderFake {
	return &ExchangeProviderFake{testbed, config, (*testbed)[0:HISTORY_SIZE], 0, map[string]float64{"USDT": 10000, "BTC": 0}, nil, nil}
}