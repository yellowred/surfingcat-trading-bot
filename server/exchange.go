package main

import (
	bittrexTicks "github.com/yellowred/golang-bittrex-api/bittrex"
	bittrexPrivate "github.com/toorop/go-bittrex"
	"time"
	"fmt"
)

type Balance struct {
	Currency      string  `json:"Currency"`
	Balance       float64 `json:"Balance"`
	Available     float64 `json:"Available"`
	Pending       float64 `json:"Pending"`
	CryptoAddress string  `json:"CryptoAddress"`
	Requested     bool    `json:"Requested"`
	Uuid          string  `json:"Uuid"`
}

// CandleStick represents a single candlestick in a chart.
type CandleStick struct {
	High       float64    `json:"H,required"`
	Open       float64    `json:"O,required"`
	Close      float64    `json:"C,required"`
	Low        float64    `json:"L,required"`
	Volume     float64    `json:"V,required"`
	BaseVolume float64    `json:"BV,required"`
	Timestamp  candleTime `json:"T,required"`
}

type candleTime time.Time

func (t *candleTime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("could not parse time %s", string(b))
	}
	// trim enclosing ""
	result, err := time.Parse("2006-01-02T15:04:05", string(b[1:len(b)-1]))
	if err != nil {
		return fmt.Errorf("could not parse time: %v", err)
	}
	*t = candleTime(result)
	return nil
}

func (t candleTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02T15:04:05"))
    return []byte(stamp), nil
}

type MarketSummary struct {
	MarketName     string  `json:"MarketName,required"`     //The name of the market (e.g. BTC-ETH).
	High           float64 `json:"High,required"`           // The 24h high for the market.
	Low            float64 `json:"Low,required"`            // The 24h low for the market.
	Last           float64 `json:"Last,required"`           // The value of the last trade for the market (in base currency).
	Bid            float64 `json:"Bid,required"`            // The current highest bid value for the market.
	Ask            float64 `json:"Ask,required"`            // The current lowest ask value for the market.
	Volume         float64 `json:"Volume,required"`         // The 24h volume of the market, in market currency.
	BaseVolume     float64 `json:"BaseVolume,required"`     // The 24h volume for the market, in base currency.
}

type ExchangeProvider interface {
	
	Balances() ([]Balance, error)

	Balance(ticker string) (Balance, error)

	Buy(ticker string, amount float64, rate float64) (string, error)

	Sell(ticker string, amount float64, rate float64) (string, error)

	Name() string

	AllCandleSticks(market string, interval string) ([]CandleStick, error)

	LastCandleStick(market string, interval string) (CandleStick, error)

	MarketSummary(market string) (MarketSummary, error)
}

const EXCHANGE_PROVIDER_BITTREX = "bittrex"

type ExchangeProviderBittrex struct {
	client *bittrexPrivate.Bittrex
	config map[string]string
}

func (p ExchangeProviderBittrex) Balances() ([]Balance, error) {
	var balances []Balance
	balancesBittrex, err := p.client.GetBalances()
	if err != nil {
		return nil, err
	}
	for _, bln := range balancesBittrex  {
		balances = append(balances, Balance{bln.Currency, bln.Balance, bln.Available, bln.Pending, bln.CryptoAddress, bln.Requested, bln.Uuid})
	}
	return balances, nil
}

func (p ExchangeProviderBittrex) Balance(ticker string) (Balance, error) {
	bln, err := p.client.GetBalance(ticker)
	if err != nil {
		return Balance{}, err
	}
	return Balance{bln.Currency, bln.Balance, bln.Available, bln.Pending, bln.CryptoAddress, bln.Requested, bln.Uuid}, nil
}

func (p ExchangeProviderBittrex) Buy(ticker string, amount float64, rate float64) (string, error) {
	return p.client.BuyLimit(ticker, amount, rate)
}

func (p ExchangeProviderBittrex) Sell(ticker string, amount float64, rate float64) (string, error) {
	return p.client.SellLimit(ticker, amount, rate)
}

func (p ExchangeProviderBittrex) Name() string {
	return EXCHANGE_PROVIDER_BITTREX
}

func (p ExchangeProviderBittrex) AllCandleSticks(market string, interval string) ([]CandleStick, error) {
	var res []CandleStick
	rBittrex, err := bittrexTicks.GetTicks(market, interval)
	if err != nil {
		return nil, err
	}
	for _, r := range rBittrex  {
		t := candleTime{}
		rtJson, err := r.Timestamp.MarshalJSON()
		if err != nil {
			return nil, err
		}
		t.UnmarshalJSON(rtJson)
		res = append(res, CandleStick{r.High, r.Open, r.Close, r.Low, r.Volume, r.BaseVolume, t})
	}
	return res, nil
}

func (p ExchangeProviderBittrex) LastCandleStick(market string, interval string) (CandleStick, error) {
	rBittrex, err := bittrexTicks.GetLatestTick(market, interval)
	if err != nil {
		return CandleStick{}, err
	}
	t := candleTime{}
	rtJson, err := rBittrex.Timestamp.MarshalJSON()
	if err != nil {
		return CandleStick{}, err
	}
	t.UnmarshalJSON(rtJson)
	return CandleStick{rBittrex.High, rBittrex.Open, rBittrex.Close, rBittrex.Low, rBittrex.Volume, rBittrex.BaseVolume, t}, nil
}

func (p ExchangeProviderBittrex) MarketSummary(market string) (MarketSummary, error) {
	r, err := bittrexTicks.GetMarketSummary(market)
	if err != nil {
		return MarketSummary{}, err
	}
	return MarketSummary{r.MarketName, r.High, r.Low, r.Last, r.Bid, r.Ask, r.Volume, r.BaseVolume}, nil
}

func ExchangeClient(name string, config map[string]string) ExchangeProvider {
	if name == EXCHANGE_PROVIDER_BITTREX {
		c := bittrexPrivate.New(BittrexApiKeys())
		return ExchangeProviderBittrex{c, config}
	} else {
		panic(fmt.Sprintf("Unknown exchange provider: %s", name))
	}
}