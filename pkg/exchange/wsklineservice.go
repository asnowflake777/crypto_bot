package exchange

import "context"

type WsKlineRequest struct {
	Symbol   string
	Interval string
}

type WsKlinesService interface {
	WsKlines(context.Context, WsKlineRequest) (<-chan *WsKlineEvent, <-chan error, error)
}

type WsKlineEvent struct {
	Event  string  `json:"e"`
	Time   int64   `json:"E"`
	Symbol string  `json:"s"`
	Kline  WsKline `json:"k"`
}

type WsKline struct {
	StartTime            int64   `json:"t"`
	EndTime              int64   `json:"T"`
	Symbol               string  `json:"s"`
	Interval             string  `json:"i"`
	FirstTradeID         int64   `json:"f"`
	LastTradeID          int64   `json:"L"`
	Open                 float64 `json:"o"`
	Close                float64 `json:"c"`
	High                 float64 `json:"h"`
	Low                  float64 `json:"l"`
	Volume               float64 `json:"v"`
	TradeNum             int64   `json:"n"`
	IsFinal              bool    `json:"x"`
	QuoteVolume          float64 `json:"q"`
	ActiveBuyVolume      float64 `json:"V"`
	ActiveBuyQuoteVolume float64 `json:"Q"`
}
