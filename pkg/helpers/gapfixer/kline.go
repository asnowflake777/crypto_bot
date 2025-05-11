package gapfixer

import (
	"context"
	"fmt"
	"time"

	"crypto_bot/pkg/exchange/models"
	"crypto_bot/pkg/storage/pgdb"
)

const MaxDelta = 1000

//go:generate mockgen -source=kline.go -destination=mocks/kline.go
type Exchange interface {
	WsKlines(context.Context, models.WsKlineRequest) (<-chan *models.WsKlineEvent, <-chan error, error)
	Klines(context.Context, models.KlinesRequest) ([]*models.Kline, error)
}

type Storage interface {
	WriteKlines(context.Context, pgdb.WriteKlinesRequest) ([]*pgdb.Kline, error)
	ReadKlines(context.Context, pgdb.ReadKlinesRequest) ([]*pgdb.Kline, error)
}

func FixGaps(ctx context.Context, ex Exchange, s Storage, symbol, interval string, from, to time.Time, chunkSize int) error {
	_, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("could not parse interval: %w", err)
	}
	var openTime, closeTime = from.UnixMilli(), to.UnixMilli()
	for i := 0; closeTime > openTime; i++ {
		klines, err := s.ReadKlines(ctx, pgdb.ReadKlinesRequest{
			Symbol:    symbol,
			Interval:  interval,
			OpenTime:  openTime,
			CloseTime: closeTime,
			Limit:     uint64(chunkSize),
		})
		if err != nil {
			return fmt.Errorf("read klines: %w", err)
		}

		gaps := findGaps(klines, openTime, closeTime, chunkSize)
		for _, g := range gaps {
			if err = fixGap(ctx, ex, s, symbol, interval, chunkSize, g); err != nil {
				return fmt.Errorf("get gap klines: %w", err)
			}

		}
		if len(klines) < chunkSize {
			return nil
		}
		openTime = klines[len(klines)-1].CloseTime + 1
	}
	return nil
}

type gap struct {
	start, end int64
}

func findGaps(klines []*pgdb.Kline, from, to int64, chunkSize int) []gap {
	if len(klines) == 0 {
		return []gap{{start: from, end: to}}
	}

	first, last := klines[0], klines[len(klines)-1]
	duration := first.CloseTime - first.OpenTime + 1
	chunkLastCloseTime := first.OpenTime + duration*int64(chunkSize) - 1
	if len(klines) == chunkSize && first.OpenTime == from && last.CloseTime == chunkLastCloseTime {
		return nil
	}

	var gaps []gap
	if first.OpenTime > from {
		gaps = append(gaps, gap{start: from, end: first.OpenTime - 1})
	}
	for i, j := 0, 1; j < len(klines); i, j = i+1, j+1 {
		left, right := klines[i], klines[j]
		if left.CloseTime+1 != right.OpenTime {
			gaps = append(gaps, gap{start: klines[i].CloseTime + 1, end: klines[j].OpenTime - 1})
		}
	}

	if len(klines) < chunkSize && last.CloseTime < to {
		gaps = append(gaps, gap{start: last.CloseTime + 1, end: to})
	}

	return gaps
}

func fixGap(ctx context.Context, ex Exchange, s Storage, symbol, interval string, limit int, g gap) error {
	for g.start < g.end {
		fmt.Printf("fixGap: start: %d, end: %d\n", g.start, g.end)
		klines, err := ex.Klines(ctx, models.KlinesRequest{
			Symbol:    symbol,
			Interval:  interval,
			Limit:     limit,
			StartTime: g.start,
			EndTime:   g.end,
		})
		if err != nil {
			return fmt.Errorf("get klines: %w", err)
		}
		if len(klines) == 0 {
			return nil
		}
		_, err = s.WriteKlines(ctx, pgdb.WriteKlinesRequest{
			Symbol:   symbol,
			Interval: interval,
			Klines:   convertKlines(klines),
		})
		if err != nil {
			return fmt.Errorf("write klines: %w", err)
		}
		if len(klines) < limit {
			return nil
		}
		g.start = klines[len(klines)-1].CloseTime + 1
	}
	return nil
}

func convertKlines(klines []*models.Kline) []*pgdb.Kline {
	converted := make([]*pgdb.Kline, 0, len(klines))
	for _, k := range klines {
		converted = append(converted, &pgdb.Kline{
			OpenTime:  k.OpenTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			CloseTime: k.CloseTime,
			TradeNum:  k.TradeNum,
		})
	}
	return converted
}
