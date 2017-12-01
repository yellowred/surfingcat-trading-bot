package exchange

import (
	"time"
	"fmt"
	configManager "github.com/yellowred/surfingcat-trading-bot/server/config"
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
	Timestamp  CandleTime `json:"T,required"`
}

type CandleTime time.Time

func (t *CandleTime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("could not parse time %s", string(b))
	}
	// trim enclosing ""
	result, err := time.Parse("2006-01-02T15:04:05", string(b[1:len(b)-1]))
	if err != nil {
		return fmt.Errorf("could not parse time: %v", err)
	}
	*t = CandleTime(result)
	return nil
}

func (t *CandleTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(*t).Format("2006-01-02T15:04:05"))
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

	Buy(market string, amount float64, rate float64) (string, error)

	Sell(market string, amount float64, rate float64) (string, error)

	Name() string

	AllCandleSticks(market string, interval string) ([]CandleStick, error)

	LastCandleStick(market string, interval string) (CandleStick, error)

	MarketSummary(market string) (MarketSummary, error)
}

const EXCHANGE_PROVIDER_BITTREX = "bittrex"
const EXCHANGE_PROVIDER_FAKE = "fake"


func ExchangeClient(name string, config map[string]string) ExchangeProvider {
	if name == EXCHANGE_PROVIDER_BITTREX {
		pbk, pvk := configManager.BittrexApiKeys()
		return NewExchangeProviderBittrex(pbk, pvk, config)
	} else {
		panic(fmt.Sprintf("Unknown exchange provider: %s", name))
	}
}