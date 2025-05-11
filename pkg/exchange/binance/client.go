package binance

import (
	"context"
	"fmt"

	"crypto_bot/pkg/exchange/models"
	"crypto_bot/pkg/exchange/utils"

	"github.com/adshao/go-binance/v2"
)

const MaxLimit = 5000

type Client struct {
	b *binance.Client
}

func NewClient(apiKey, secretKey string) *Client {
	return &Client{b: binance.NewClient(apiKey, secretKey)}
}

func (c Client) Klines(ctx context.Context, r models.KlinesRequest) (res []*models.Kline, err error) {
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
		if r.Limit > MaxLimit {
			return nil, fmt.Errorf("limit exceeded")
		}
		s = s.Limit(r.Limit)
	}
	extKlines, err := s.Do(ctx)
	if err != nil {
		return nil, err
	}
	klines := make([]*models.Kline, len(extKlines))
	for i, kline := range extKlines {
		klines[i] = utils.FromExtKlineToInt(kline)
	}
	return klines, nil
}

func errHandler(errs chan error) func(error) {
	return func(err error) {
		errs <- err
	}
}

func eventHandler(events chan *models.WsKlineEvent) func(*binance.WsKlineEvent) {
	return func(event *binance.WsKlineEvent) {
		events <- utils.FromExtWsKlineEventToInt(event)
	}
}

func (c Client) WsKlines(ctx context.Context, r models.WsKlineRequest) (<-chan *models.WsKlineEvent, <-chan error, error) {
	errs := make(chan error, 100)
	events := make(chan *models.WsKlineEvent, 100)

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

func (c Client) CreateOrder(ctx context.Context, r models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	order, err := c.b.NewCreateOrderService().
		Symbol(r.Symbol).
		Side(binance.SideType(r.Side)).
		Type(binance.OrderType(r.Type)).
		TimeInForce(binance.TimeInForceType(r.InTimeForce)).
		Quantity(utils.Float2str(r.Quantity)).
		Price(utils.Float2str(r.Price)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	fills := make([]*models.Fill, len(order.Fills))
	for i, fill := range order.Fills {
		fills[i] = &models.Fill{
			TradeID:         fill.TradeID,
			Price:           utils.Str2float(fill.Price),
			Quantity:        utils.Str2float(fill.Quantity),
			Commission:      utils.Str2float(fill.Commission),
			CommissionAsset: utils.Str2float(fill.CommissionAsset),
		}
	}
	return &models.CreateOrderResponse{
		Symbol:                   order.Symbol,
		OrderID:                  order.OrderID,
		ClientOrderID:            order.ClientOrderID,
		TransactTime:             order.TransactTime,
		Price:                    utils.Str2float(order.Price),
		OrigQuantity:             utils.Str2float(order.OrigQuantity),
		ExecutedQuantity:         utils.Str2float(order.ExecutedQuantity),
		CummulativeQuoteQuantity: utils.Str2float(order.CummulativeQuoteQuantity),
		IsIsolated:               order.IsIsolated,
		Status:                   models.OrderStatusType(order.Status),
		TimeInForce:              models.TimeInForceType(order.TimeInForce),
		Type:                     models.OrderType(order.Type),
		Side:                     models.SideType(order.Side),
		Fills:                    fills,
		MarginBuyBorrowAmount:    utils.Str2float(order.MarginBuyBorrowAmount),
		MarginBuyBorrowAsset:     utils.Str2float(order.MarginBuyBorrowAsset),
		SelfTradePreventionMode:  models.SelfTradePreventionMode(order.SelfTradePreventionMode),
	}, nil
}

func (c Client) GetOrder(ctx context.Context, r models.ReadOrderRequest) (*models.Order, error) {
	o, err := c.b.NewGetOrderService().Symbol(r.Symbol).OrderID(r.ID).Do(ctx)
	if err != nil {
		return nil, err
	}
	return utils.FromExtOrderToInt(o), nil
}

func (c Client) CancelOrder(ctx context.Context, r models.CancelOrderRequest) (*models.CancelOrderResponse, error) {
	o, err := c.b.NewCancelOrderService().Symbol(r.Symbol).OrderID(r.ID).Do(ctx)
	if err != nil {
		return nil, err
	}
	return &models.CancelOrderResponse{
		Symbol:                   o.Symbol,
		OrigClientOrderID:        o.OrigClientOrderID,
		OrderID:                  o.OrderID,
		OrderListID:              o.OrderListID,
		ClientOrderID:            o.ClientOrderID,
		TransactTime:             o.TransactTime,
		Price:                    utils.Str2float(o.Price),
		OrigQuantity:             utils.Str2float(o.OrigQuantity),
		ExecutedQuantity:         utils.Str2float(o.ExecutedQuantity),
		CummulativeQuoteQuantity: utils.Str2float(o.CummulativeQuoteQuantity),
		Status:                   models.OrderStatusType(o.Status),
		TimeInForce:              models.TimeInForceType(o.TimeInForce),
		Type:                     models.OrderType(o.Type),
		Side:                     models.SideType(o.Side),
		SelfTradePreventionMode:  models.SelfTradePreventionMode(o.SelfTradePreventionMode),
	}, nil
}

func (c Client) ListOrders(ctx context.Context, r models.ListOrdersRequest) ([]*models.Order, error) {
	orders, err := c.b.NewListOrdersService().Symbol(r.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*models.Order, len(orders))
	for i, o := range orders {
		res[i] = utils.FromExtOrderToInt(o)
	}
	return res, err
}

func (c Client) ListOpenOrders(ctx context.Context, r models.ListOpenOrdersRequest) ([]*models.Order, error) {
	openOrders, err := c.b.NewListOpenOrdersService().Symbol(r.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*models.Order, len(openOrders))
	for i, o := range openOrders {
		res[i] = utils.FromExtOrderToInt(o)
	}
	return res, err
}

func (c Client) GetAccount(ctx context.Context) (*models.Account, error) {
	acc, err := c.b.NewGetAccountService().OmitZeroBalances(true).Do(ctx)
	if err != nil {
		return nil, err
	}
	balances := make([]models.Balance, len(acc.Balances))
	for i, b := range acc.Balances {
		balances[i] = models.Balance(b)
	}
	return &models.Account{
		MakerCommission:  acc.MakerCommission,
		TakerCommission:  acc.TakerCommission,
		BuyerCommission:  acc.BuyerCommission,
		SellerCommission: acc.SellerCommission,
		CommissionRates:  models.CommissionRates(acc.CommissionRates),
		CanTrade:         acc.CanTrade,
		CanWithdraw:      acc.CanWithdraw,
		CanDeposit:       acc.CanDeposit,
		UpdateTime:       acc.UpdateTime,
		AccountType:      acc.AccountType,
		Balances:         balances,
		Permissions:      acc.Permissions,
		UID:              acc.UID,
	}, nil
}
