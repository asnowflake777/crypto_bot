package exchange

import "context"

type KlinesService interface {
	Klines(ctx context.Context, request KlinesRequest) (res []*Kline, err error)
}

type Kline struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	TradeNum                 int64  `json:"tradeNum"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
}

type KlinesRequest struct {
	Symbol    string
	Interval  string
	Limit     int
	StartTime int64
	EndTime   int64
}
