package binance

import (
	"crypto_bot/pkg/exchange"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"strconv"
)

func FromExtKlineToInt(kline *binance.Kline) *exchange.Kline {
	return &exchange.Kline{
		OpenTime:                 kline.OpenTime,
		Open:                     str2float(kline.Open),
		High:                     str2float(kline.High),
		Low:                      str2float(kline.Low),
		Close:                    str2float(kline.Close),
		Volume:                   str2float(kline.Volume),
		CloseTime:                kline.CloseTime,
		QuoteAssetVolume:         str2float(kline.QuoteAssetVolume),
		TradeNum:                 kline.TradeNum,
		TakerBuyBaseAssetVolume:  str2float(kline.TakerBuyBaseAssetVolume),
		TakerBuyQuoteAssetVolume: str2float(kline.TakerBuyQuoteAssetVolume),
	}
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
	return exchange.WsKline{
		StartTime:            kline.StartTime,
		EndTime:              kline.EndTime,
		Symbol:               kline.Symbol,
		Interval:             kline.Interval,
		FirstTradeID:         kline.FirstTradeID,
		LastTradeID:          kline.LastTradeID,
		Open:                 str2float(kline.Open),
		Close:                str2float(kline.Close),
		High:                 str2float(kline.High),
		Low:                  str2float(kline.Low),
		Volume:               str2float(kline.Volume),
		TradeNum:             kline.TradeNum,
		IsFinal:              kline.IsFinal,
		QuoteVolume:          str2float(kline.QuoteVolume),
		ActiveBuyVolume:      str2float(kline.ActiveBuyVolume),
		ActiveBuyQuoteVolume: str2float(kline.ActiveBuyQuoteVolume),
	}
}

func FromExtOrderToInt(o *binance.Order) *exchange.Order {
	return &exchange.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		OrderListId:              o.OrderListId,
		ClientOrderID:            o.ClientOrderID,
		Price:                    str2float(o.Price),
		OrigQuantity:             str2float(o.OrigQuantity),
		ExecutedQuantity:         str2float(o.ExecutedQuantity),
		CummulativeQuoteQuantity: str2float(o.CummulativeQuoteQuantity),
		Status:                   exchange.OrderStatusType(o.Status),
		TimeInForce:              exchange.TimeInForceType(o.TimeInForce),
		Type:                     exchange.OrderType(o.Type),
		Side:                     exchange.SideType(o.Side),
		StopPrice:                str2float(o.StopPrice),
		IcebergQuantity:          str2float(o.IcebergQuantity),
		Time:                     o.Time,
		UpdateTime:               o.UpdateTime,
		IsWorking:                o.IsWorking,
		IsIsolated:               o.IsIsolated,
		OrigQuoteOrderQuantity:   str2float(o.OrigQuoteOrderQuantity),
	}
}

func str2float(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to convert string to float: %s", err))
	}
	return f
}

func float2str(f float64) string {
	return fmt.Sprintf("%.8f", f)
}
