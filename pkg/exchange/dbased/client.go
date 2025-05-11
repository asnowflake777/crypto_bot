package dbased

import (
	"context"
	"fmt"
	"time"

	"crypto_bot/pkg/exchange/models"
	"crypto_bot/pkg/exchange/utils"
	"crypto_bot/pkg/storage/pgdb"
)

type Client struct {
	s         *pgdb.Client
	user      *pgdb.User
	startTime int64
}

func NewClient(ctx context.Context, s *pgdb.Client, username string, startTime int64) (*Client, error) {
	user, err := s.ReadUser(ctx, pgdb.ReadUserRequest{Login: username})
	if err != nil {
		return nil, err
	}
	return &Client{s: s, user: user, startTime: startTime}, err
}

func (c *Client) Klines(ctx context.Context, r models.KlinesRequest) ([]*models.Kline, error) {
	klines, err := c.s.ReadKlines(ctx, pgdb.ReadKlinesRequest{
		Symbol:    r.Symbol,
		Interval:  r.Interval,
		OpenTime:  r.StartTime,
		CloseTime: r.EndTime,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*models.Kline, len(klines))
	for i, k := range klines {
		res[i] = &models.Kline{
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

func (c *Client) WsKlines(ctx context.Context, r models.WsKlineRequest) (<-chan *models.WsKlineEvent, <-chan error, error) {
	klines, err := c.Klines(ctx, models.KlinesRequest{
		Symbol:    r.Symbol,
		Interval:  r.Interval,
		StartTime: c.startTime,
	})
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan *models.WsKlineEvent)
	errs := make(chan error)
	go func() {
		defer close(ch)
		defer close(errs)
		for _, k := range klines {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- &models.WsKlineEvent{
					Event:  "pgdb",
					Time:   k.OpenTime,
					Symbol: r.Symbol,
					Kline: models.WsKline{
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

func (c *Client) CreateOrder(ctx context.Context, r models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
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
	return &models.CreateOrderResponse{
		Symbol:                   order.Symbol,
		OrderID:                  order.ID,
		TransactTime:             time.Now().UnixMilli(),
		Price:                    order.Price,
		OrigQuantity:             order.Quantity,
		ExecutedQuantity:         order.Quantity,
		CummulativeQuoteQuantity: order.Quantity,
		IsIsolated:               false,
		Status:                   models.OrderStatusTypeFilled,
		TimeInForce:              models.TimeInForceTypeGTC,
		Type:                     models.OrderTypeMarket,
		Side:                     r.Side,
		Fills: []*models.Fill{{
			TradeID:         order.ID,
			Price:           order.Price,
			Quantity:        order.Quantity,
			Commission:      0.1,
			CommissionAsset: 0.1,
		}},
	}, nil
}

func (c *Client) GetOrder(ctx context.Context, r models.ReadOrderRequest) (*models.Order, error) {
	o, err := c.s.ReadOrder(ctx, pgdb.ReadOrderRequest{ID: r.ID})
	if err != nil {
		return nil, err
	}
	return &models.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.ID,
		Price:                    o.Price,
		OrigQuantity:             o.Quantity,
		ExecutedQuantity:         o.Quantity,
		CummulativeQuoteQuantity: o.Quantity,
		Status:                   models.OrderStatusTypeFilled,
		TimeInForce:              models.TimeInForceTypeGTC,
		Type:                     models.OrderType(o.Type),
		Side:                     models.SideType(o.Side),
		StopPrice:                o.Price,
		OrigQuoteOrderQuantity:   o.Quantity,
	}, nil
}

func (c *Client) CancelOrder(_ context.Context, _ models.CancelOrderRequest) (*models.CancelOrderResponse, error) {
	return nil, fmt.Errorf("not found")
}

func (c *Client) ListOrders(ctx context.Context, r models.ListOrdersRequest) ([]*models.Order, error) {
	orders, err := c.s.ReadOrders(ctx, pgdb.ReadOrdersRequest{UserUID: c.user.UID, Symbol: r.Symbol})
	if err != nil {
		return nil, err
	}
	res := make([]*models.Order, len(orders))
	for i, o := range orders {
		res[i] = &models.Order{
			Symbol:                   o.Symbol,
			OrderID:                  o.ID,
			Price:                    o.Price,
			OrigQuantity:             o.Quantity,
			ExecutedQuantity:         o.Quantity,
			CummulativeQuoteQuantity: o.Quantity,
			Status:                   models.OrderStatusTypeFilled,
			TimeInForce:              models.TimeInForceTypeGTC,
			Type:                     models.OrderType(o.Type),
			Side:                     models.SideType(o.Side),
			StopPrice:                o.Price,
			OrigQuoteOrderQuantity:   o.Quantity,
		}
	}
	return res, nil
}

func (c *Client) ListOpenOrders(_ context.Context, _ models.ListOpenOrdersRequest) ([]*models.Order, error) {
	return nil, nil
}

func (c *Client) GetAccount(ctx context.Context) (*models.Account, error) {
	balances, err := c.s.ReadBalances(ctx, pgdb.ReadBalancesRequest{UserUID: c.user.UID})
	if err != nil {
		return nil, err
	}
	exchangeBalances := make([]models.Balance, len(balances))
	for i, b := range balances {
		exchangeBalances[i] = models.Balance{
			Asset:  b.Asset,
			Free:   utils.Float2str(b.Free),
			Locked: utils.Float2str(b.Locked),
		}
	}
	return &models.Account{
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
