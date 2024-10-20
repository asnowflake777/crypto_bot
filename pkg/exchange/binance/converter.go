package binance

import (
	"crypto_bot/pkg/exchange"
	"github.com/adshao/go-binance/v2"
)

func FromExtKlineToInt(kline *binance.Kline) *exchange.Kline {
	return (*exchange.Kline)(kline)
}

func FromExtWsKlineEventToInt(event *binance.WsKlineEvent) *exchange.WsKlineEvent {
	if event == nil {
		return nil
	}
	return &exchange.WsKlineEvent{
		Event:  event.Event,
		Time:   event.Time,
		Symbol: event.Symbol,
		Kline:  FromExtWsKlineToInt(event.Kline),
	}
}

func FromExtWsKlineToInt(kline binance.WsKline) exchange.WsKline {
	return exchange.WsKline(kline)
}
