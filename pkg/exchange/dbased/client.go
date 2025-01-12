package dbased

import (
	"context"
	"fmt"
	"time"

	"crypto_bot/pkg/exchange"
	"crypto_bot/pkg/exchange/utils"
	"crypto_bot/pkg/storage/pgdb"
)

type Client struct {
	s         *pgdb.Client
	user      *pgdb.User
	startTime int64
}

var _ exchange.Client = (*Client)(nil)

func NewClient(ctx context.Context, s *pgdb.Client, username string, startTime int64) (*Client, error) {
	user, err := s.ReadUser(ctx, pgdb.ReadUserRequest{Login: username})
	if err != nil {
		return nil, err
	}
	return &Client{s: s, user: user, startTime: startTime}, err
}

func (c *Client) Klines(ctx context.Context, r exchange.KlinesRequest) ([]*exchange.Kline, error) {
	klines, err := c.s.ReadKlines(ctx, pgdb.ReadKlinesRequest{
		Symbol:    r.Symbol,
		Interval:  r.Interval,
		OpenTime:  r.StartTime,
		CloseTime: r.EndTime,
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

func (c *Client) WsKlines(ctx context.Context, r exchange.WsKlineRequest) (<-chan *exchange.WsKlineEvent, <-chan error, error) {
	klines, err := c.Klines(ctx, exchange.KlinesRequest{
		Symbol:    r.Symbol,
		Interval:  r.Interval,
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
					Symbol: r.Symbol,
					Kline: exchange.WsKline{
						StartTime: k.OpenTime,
						EndTime:   k.CloseTime,
						Symbol:    r.Symbol,
						Interval:  r.Interval,
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

func (c *Client) CreateOrder(ctx context.Context, r exchange.CreateOrderRequest) (*exchange.CreateOrderResponse, error) {
	order, err := c.s.CreateOrder(ctx, pgdb.CreateOrderRequest{
		Symbol:   r.Symbol,
		Price:    r.Price,
		Quantity: r.Quantity,
		Type:     string(r.Type),
		Side:     string(r.Side),
	})
	if err != nil {
		return nil, err
	}
	return &exchange.CreateOrderResponse{
		Symbol:                   order.Symbol,
		OrderID:                  order.ID,
		TransactTime:             time.Now().UnixMilli(),
		Price:                    order.Price,
		OrigQuantity:             order.Quantity,
		ExecutedQuantity:         order.Quantity,
		CummulativeQuoteQuantity: order.Quantity,
		IsIsolated:               false,
		Status:                   exchange.OrderStatusTypeFilled,
		TimeInForce:              exchange.TimeInForceTypeGTC,
		Type:                     exchange.OrderTypeMarket,
		Side:                     r.Side,
		Fills: []*exchange.Fill{{
			TradeID:         order.ID,
			Price:           order.Price,
			Quantity:        order.Quantity,
			Commission:      0.1,
			CommissionAsset: 0.1,
		}},
	}, nil
}

func (c *Client) GetOrder(ctx context.Context, r exchange.ReadOrderRequest) (*exchange.Order, error) {
	o, err := c.s.ReadOrder(ctx, pgdb.ReadOrderRequest{ID: r.ID})
	if err != nil {
		return nil, err
	}
	return &exchange.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.ID,
		Price:                    o.Price,
		OrigQuantity:             o.Quantity,
		ExecutedQuantity:         o.Quantity,
		CummulativeQuoteQuantity: o.Quantity,
		Status:                   exchange.OrderStatusTypeFilled,
		TimeInForce:              exchange.TimeInForceTypeGTC,
		Type:                     exchange.OrderType(o.Type),
		Side:                     exchange.SideType(o.Side),
		StopPrice:                o.Price,
		OrigQuoteOrderQuantity:   o.Quantity,
	}, nil
}

func (c *Client) CancelOrder(_ context.Context, _ exchange.CancelOrderRequest) (*exchange.CancelOrderResponse, error) {
	return nil, fmt.Errorf("not found")
}

func (c *Client) ListOrders(ctx context.Context, r exchange.ListOrdersRequest) ([]*exchange.Order, error) {
	orders, err := c.s.ReadOrders(ctx, pgdb.ReadOrdersRequest{UserUID: c.user.UID, Symbol: r.Symbol})
	if err != nil {
		return nil, err
	}
	res := make([]*exchange.Order, len(orders))
	for i, o := range orders {
		res[i] = &exchange.Order{
			Symbol:                   o.Symbol,
			OrderID:                  o.ID,
			Price:                    o.Price,
			OrigQuantity:             o.Quantity,
			ExecutedQuantity:         o.Quantity,
			CummulativeQuoteQuantity: o.Quantity,
			Status:                   exchange.OrderStatusTypeFilled,
			TimeInForce:              exchange.TimeInForceTypeGTC,
			Type:                     exchange.OrderType(o.Type),
			Side:                     exchange.SideType(o.Side),
			StopPrice:                o.Price,
			OrigQuoteOrderQuantity:   o.Quantity,
		}
	}
	return res, nil
}

func (c *Client) ListOpenOrders(_ context.Context, _ exchange.ListOpenOrdersRequest) ([]*exchange.Order, error) {
	return nil, nil
}

func (c *Client) GetAccount(ctx context.Context) (*exchange.Account, error) {
	balances, err := c.s.ReadBalances(ctx, pgdb.ReadBalancesRequest{UserUID: c.user.UID})
	if err != nil {
		return nil, err
	}
	exchangeBalances := make([]exchange.Balance, len(balances))
	for i, b := range balances {
		exchangeBalances[i] = exchange.Balance{
			Asset:  b.Asset,
			Free:   utils.Float2str(b.Free),
			Locked: utils.Float2str(b.Locked),
		}
	}
	return &exchange.Account{
		CanTrade:    true,
		CanWithdraw: true,
		CanDeposit:  true,
		Balances:    exchangeBalances,
		UID:         c.user.UID,
	}, nil
}

func (c *Client) SetStartTime(startTime int64) {
	c.startTime = startTime
}

func (c *Client) SetUsername(ctx context.Context, username string) error {
	user, err := c.s.ReadUser(ctx, pgdb.ReadUserRequest{Login: username})
	if err != nil {
		return err
	}
	c.user = user
	return nil
}
