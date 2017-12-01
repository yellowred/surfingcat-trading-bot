package exchange

import (
	bittrexTicks "github.com/thebotguys/golang-bittrex-api/bittrex"
	bittrexPrivate "github.com/toorop/go-bittrex"
	"time"
)

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
		if bln.Balance > 0 || p.config["hide_zero_balances"] != "Y" {
			balances = append(balances, Balance{bln.Currency, bln.Balance, bln.Available, bln.Pending, bln.CryptoAddress, bln.Requested, bln.Uuid})
		}
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

func (p ExchangeProviderBittrex) Buy(market string, amount float64, rate float64) (string, error) {
	return p.client.BuyLimit(market, amount, rate)
}

func (p ExchangeProviderBittrex) Sell(market string, amount float64, rate float64) (string, error) {
	return p.client.SellLimit(market, amount, rate)
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
		t := CandleTime{}
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
	t := CandleTime{}
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
	// sometimes bittrex just returns an arbitrary cached market
	if r.MarketName != market {
		time.Sleep(1e3)
		return p.MarketSummary(market)
	}
	return MarketSummary{r.MarketName, r.High, r.Low, r.Last, r.Bid, r.Ask, r.Volume, r.BaseVolume}, nil
}


func NewExchangeProviderBittrex(pbk string, pvk string, config map[string]string) ExchangeProvider {
	c := bittrexPrivate.New(pbk, pvk)
	return ExchangeProviderBittrex{c, config}
}