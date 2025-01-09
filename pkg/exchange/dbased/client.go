package dbased

import (
	"context"
	"crypto_bot/pkg/exchange"
	"crypto_bot/pkg/storage/pgdb"
)

type Client struct {
	s         *pgdb.Client
	startTime int64
}

var _ exchange.Client = (*Client)(nil)

func NewClient(s *pgdb.Client, startTime int64) *Client {
	return &Client{s: s, startTime: startTime}
}

func (c *Client) Klines(ctx context.Context, req exchange.KlinesRequest) ([]*exchange.Kline, error) {
	klines, err := c.s.ReadKlines(ctx, pgdb.ReadKlinesRequest{
		Symbol:    req.Symbol,
		Interval:  req.Interval,
		OpenTime:  req.StartTime,
		CloseTime: req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*exchange.Kline, len(klines))
	for i, k := range klines {
		res[i] = &exchange.Kline{
			OpenTime:  k.OpenTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			CloseTime: k.CloseTime,
			TradeNum:  k.TradeNum,
		}
	}
	return res, nil
}

func (c *Client) WsKlines(ctx context.Context, req exchange.WsKlineRequest) (<-chan *exchange.WsKlineEvent, <-chan error, error) {
	klines, err := c.Klines(ctx, exchange.KlinesRequest{
		Symbol:    req.Symbol,
		Interval:  req.Interval,
		StartTime: c.startTime,
	})
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan *exchange.WsKlineEvent)
	errs := make(chan error)
	go func() {
		defer close(ch)
		defer close(errs)
		for _, k := range klines {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- &exchange.WsKlineEvent{
					Event:  "pgdb",
					Time:   k.OpenTime,
					Symbol: req.Symbol,
					Kline: exchange.WsKline{
						StartTime: k.OpenTime,
						EndTime:   k.CloseTime,
						Symbol:    req.Symbol,
						Interval:  req.Interval,
						Open:      k.Open,
						Close:     k.Close,
						High:      k.High,
						Low:       k.Low,
						Volume:    k.Volume,
						TradeNum:  k.TradeNum,
						IsFinal:   true,
					},
				}
			}
		}

	}()
	return ch, errs, nil
}

func (c *Client) CreateOrder(ctx context.Context, request exchange.CreateOrderRequest) (*exchange.CreateOrderResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetOrder(ctx context.Context, request exchange.ReadOrderRequest) (*exchange.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) CancelOrder(ctx context.Context, request exchange.CancelOrderRequest) (*exchange.CancelOrderResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) ListOrders(ctx context.Context, request exchange.ListOrdersRequest) ([]*exchange.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) ListOpenOrders(ctx context.Context, request exchange.ListOpenOrdersRequest) ([]*exchange.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetAccount(ctx context.Context) (*exchange.Account, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetStartTime(startTime int64) {
	c.startTime = startTime
}
