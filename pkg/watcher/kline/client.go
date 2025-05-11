package kline

import (
	"context"
	"errors"
	"log"

	"crypto_bot/pkg/exchange/models"
	"crypto_bot/pkg/storage/pgdb"
)

type Watcher struct {
	db        Storage
	ex        Exchange
	symbol    string
	interval  string
	chunkSize int
	debug     bool

	errHandler func(err error)
}

type Exchange interface {
	WsKlines(context.Context, models.WsKlineRequest) (<-chan *models.WsKlineEvent, <-chan error, error)
}

type Storage interface {
	WriteKlines(context.Context, pgdb.WriteKlinesRequest) ([]*pgdb.Kline, error)
}

func NewWatcher(ex Exchange, db Storage) *Watcher {
	return &Watcher{
		db:        db,
		ex:        ex,
		symbol:    "BTCUSDT",
		interval:  "1h",
		chunkSize: 10,
		errHandler: func(err error) {
			log.Println(err)
		},
	}
}

func (w *Watcher) SetSymbols(symbol string) *Watcher {
	w.symbol = symbol
	return w
}

func (w *Watcher) SetInterval(interval string) *Watcher {
	w.interval = interval
	return w
}

func (w *Watcher) SetChunkSize(size int) *Watcher {
	w.chunkSize = size
	return w
}

func (w *Watcher) SetErrorHandler(handler func(error)) *Watcher {
	w.errHandler = handler
	return w
}

func (w *Watcher) SetDebug(v bool) *Watcher {
	w.debug = v
	return w
}

func (w *Watcher) Start(ctx context.Context) error {
	events, errs, err := w.ex.WsKlines(ctx, models.WsKlineRequest{Symbol: w.symbol, Interval: w.interval})
	if err != nil {
		return err
	}

	go func() {
		for e := range errs {
			w.errHandler(e)
		}
	}()

	for {
		if err = w.processChunk(ctx, events); err != nil {
			return err
		}
	}
}

func (w *Watcher) processChunk(ctx context.Context, events <-chan *models.WsKlineEvent) (err error) {
	klinesToWrite := make([]*pgdb.Kline, 0, w.chunkSize)
	defer func() {
		if err != nil && len(klinesToWrite) == 0 {
			return
		}

		_, writeErr := w.db.WriteKlines(ctx, pgdb.WriteKlinesRequest{
			Symbol:   w.symbol,
			Interval: w.interval,
			Klines:   klinesToWrite,
		})
		if writeErr != nil {
			err = errors.Join(err, writeErr)
			return
		}
		if w.debug {
			log.Printf("wrote %d klines", len(klinesToWrite))
		}
	}()

	for event := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !event.Kline.IsFinal {
				continue
			}
			klinesToWrite = append(klinesToWrite, &pgdb.Kline{
				OpenTime:  event.Kline.StartTime,
				Open:      event.Kline.Open,
				High:      event.Kline.High,
				Low:       event.Kline.Low,
				Close:     event.Kline.Close,
				Volume:    event.Kline.Volume,
				CloseTime: event.Kline.EndTime,
				TradeNum:  event.Kline.TradeNum,
			})
			if len(klinesToWrite) >= w.chunkSize {
				return nil
			}
		}
	}
	return nil
}
