package binance

import (
	"context"
	"crypto_bot/pkg/exchange"
	"github.com/adshao/go-binance/v2"
)

type Client struct {
	b *binance.Client
}

var _ exchange.Client = (*Client)(nil)

func NewClient(apiKey, secretKey string) *Client {
	return &Client{b: binance.NewClient(apiKey, secretKey)}
}

func (c Client) Klines(ctx context.Context, r exchange.KlinesRequest) (res []*exchange.Kline, err error) {
	s := c.b.NewKlinesService().
		Symbol(r.Symbol).
		Interval(r.Interval)

	if r.StartTime > 0 {
		s = s.StartTime(r.StartTime)
	}
	if r.EndTime > 0 {
		s = s.EndTime(r.EndTime)
	}
	if r.Limit > 0 {
		s = s.Limit(r.Limit)
	}
	extKlines, err := s.Do(ctx)
	if err != nil {
		return nil, err
	}
	klines := make([]*exchange.Kline, len(extKlines))
	for i, kline := range extKlines {
		klines[i] = FromExtKlineToInt(kline)
	}
	return klines, nil
}

func errHandler(errs chan error) func(error) {
	return func(err error) {
		errs <- err
	}
}

func eventHandler(events chan *exchange.WsKlineEvent) func(*binance.WsKlineEvent) {
	return func(event *binance.WsKlineEvent) {
		events <- FromExtWsKlineEventToInt(event)
	}
}

func (c Client) WsKlines(ctx context.Context, r exchange.WsKlineRequest) (<-chan *exchange.WsKlineEvent, <-chan error, error) {
	errs := make(chan error, 100)
	events := make(chan *exchange.WsKlineEvent, 100)

	closeChans := func() {
		close(errs)
		close(events)
	}

	done, stop, err := binance.WsKlineServe(r.Symbol, r.Interval, eventHandler(events), errHandler(errs))
	if err != nil {
		closeChans()
		return nil, nil, err
	}

	go func() {
		defer closeChans()
		select {
		case <-done:
		case <-ctx.Done():
			stop <- struct{}{}
		}
	}()

	return events, errs, nil
}
