package utils

import (
	"fmt"
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"

	"crypto_bot/pkg/exchange/models"
)

func FromExtKlineToInt(kline *binance.Kline) *models.Kline {
	return &models.Kline{
		OpenTime:                 kline.OpenTime,
		Open:                     Str2float(kline.Open),
		High:                     Str2float(kline.High),
		Low:                      Str2float(kline.Low),
		Close:                    Str2float(kline.Close),
		Volume:                   Str2float(kline.Volume),
		CloseTime:                kline.CloseTime,
		QuoteAssetVolume:         Str2float(kline.QuoteAssetVolume),
		TradeNum:                 kline.TradeNum,
		TakerBuyBaseAssetVolume:  Str2float(kline.TakerBuyBaseAssetVolume),
		TakerBuyQuoteAssetVolume: Str2float(kline.TakerBuyQuoteAssetVolume),
	}
}

func FromExtWsKlineEventToInt(event *binance.WsKlineEvent) *models.WsKlineEvent {
	if event == nil {
		return nil
	}
	return &models.WsKlineEvent{
		Event:  event.Event,
		Time:   event.Time,
		Symbol: event.Symbol,
		Kline:  FromExtWsKlineToInt(event.Kline),
	}
}

func FromExtWsKlineToInt(kline binance.WsKline) models.WsKline {
	return models.WsKline{
		StartTime:            kline.StartTime,
		EndTime:              kline.EndTime,
		Symbol:               kline.Symbol,
		Interval:             kline.Interval,
		FirstTradeID:         kline.FirstTradeID,
		LastTradeID:          kline.LastTradeID,
		Open:                 Str2float(kline.Open),
		Close:                Str2float(kline.Close),
		High:                 Str2float(kline.High),
		Low:                  Str2float(kline.Low),
		Volume:               Str2float(kline.Volume),
		TradeNum:             kline.TradeNum,
		IsFinal:              kline.IsFinal,
		QuoteVolume:          Str2float(kline.QuoteVolume),
		ActiveBuyVolume:      Str2float(kline.ActiveBuyVolume),
		ActiveBuyQuoteVolume: Str2float(kline.ActiveBuyQuoteVolume),
	}
}

func FromExtOrderToInt(o *binance.Order) *models.Order {
	return &models.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		OrderListId:              o.OrderListId,
		ClientOrderID:            o.ClientOrderID,
		Price:                    Str2float(o.Price),
		OrigQuantity:             Str2float(o.OrigQuantity),
		ExecutedQuantity:         Str2float(o.ExecutedQuantity),
		CummulativeQuoteQuantity: Str2float(o.CummulativeQuoteQuantity),
		Status:                   models.OrderStatusType(o.Status),
		TimeInForce:              models.TimeInForceType(o.TimeInForce),
		Type:                     models.OrderType(o.Type),
		Side:                     models.SideType(o.Side),
		StopPrice:                Str2float(o.StopPrice),
		IcebergQuantity:          Str2float(o.IcebergQuantity),
		Time:                     o.Time,
		UpdateTime:               o.UpdateTime,
		IsWorking:                o.IsWorking,
		IsIsolated:               o.IsIsolated,
		OrigQuoteOrderQuantity:   Str2float(o.OrigQuoteOrderQuantity),
	}
}

func Str2float(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to convert string to float: %s", err))
	}
	return f
}

func Float2str(f float64) string {
	return fmt.Sprintf("%.8f", f)
}
