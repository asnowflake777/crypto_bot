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

func (c Client) CreateOrder(ctx context.Context, r exchange.CreateOrderRequest) (*exchange.CreateOrderResponse, error) {
	order, err := c.b.NewCreateOrderService().
		Symbol(r.Symbol).
		Side(binance.SideType(r.Side)).
		Type(binance.OrderType(r.Type)).
		TimeInForce(binance.TimeInForceType(r.InTimeForce)).
		Quantity(float2str(r.Quantity)).
		Price(float2str(r.Price)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	fills := make([]*exchange.Fill, len(order.Fills))
	for i, fill := range order.Fills {
		fills[i] = &exchange.Fill{
			TradeID:         fill.TradeID,
			Price:           str2float(fill.Price),
			Quantity:        str2float(fill.Quantity),
			Commission:      str2float(fill.Commission),
			CommissionAsset: str2float(fill.CommissionAsset),
		}
	}
	return &exchange.CreateOrderResponse{
		Symbol:                   order.Symbol,
		OrderID:                  order.OrderID,
		ClientOrderID:            order.ClientOrderID,
		TransactTime:             order.TransactTime,
		Price:                    str2float(order.Price),
		OrigQuantity:             str2float(order.OrigQuantity),
		ExecutedQuantity:         str2float(order.ExecutedQuantity),
		CummulativeQuoteQuantity: str2float(order.CummulativeQuoteQuantity),
		IsIsolated:               order.IsIsolated,
		Status:                   exchange.OrderStatusType(order.Status),
		TimeInForce:              exchange.TimeInForceType(order.TimeInForce),
		Type:                     exchange.OrderType(order.Type),
		Side:                     exchange.SideType(order.Side),
		Fills:                    fills,
		MarginBuyBorrowAmount:    str2float(order.MarginBuyBorrowAmount),
		MarginBuyBorrowAsset:     str2float(order.MarginBuyBorrowAsset),
		SelfTradePreventionMode:  exchange.SelfTradePreventionMode(order.SelfTradePreventionMode),
	}, nil
}

func (c Client) GetOrder(ctx context.Context, r exchange.ReadOrderRequest) (*exchange.Order, error) {
	o, err := c.b.NewGetOrderService().Symbol(r.Symbol).OrderID(r.ID).Do(ctx)
	if err != nil {
		return nil, err
	}
	return FromExtOrderToInt(o), nil
}

func (c Client) CancelOrder(ctx context.Context, r exchange.CancelOrderRequest) (*exchange.CancelOrderResponse, error) {
	o, err := c.b.NewCancelOrderService().Symbol(r.Symbol).OrderID(r.ID).Do(ctx)
	if err != nil {
		return nil, err
	}
	return &exchange.CancelOrderResponse{
		Symbol:                   o.Symbol,
		OrigClientOrderID:        o.OrigClientOrderID,
		OrderID:                  o.OrderID,
		OrderListID:              o.OrderListID,
		ClientOrderID:            o.ClientOrderID,
		TransactTime:             o.TransactTime,
		Price:                    str2float(o.Price),
		OrigQuantity:             str2float(o.OrigQuantity),
		ExecutedQuantity:         str2float(o.ExecutedQuantity),
		CummulativeQuoteQuantity: str2float(o.CummulativeQuoteQuantity),
		Status:                   exchange.OrderStatusType(o.Status),
		TimeInForce:              exchange.TimeInForceType(o.TimeInForce),
		Type:                     exchange.OrderType(o.Type),
		Side:                     exchange.SideType(o.Side),
		SelfTradePreventionMode:  exchange.SelfTradePreventionMode(o.SelfTradePreventionMode),
	}, nil
}

func (c Client) ListOrders(ctx context.Context, r exchange.ListOrdersRequest) ([]*exchange.Order, error) {
	orders, err := c.b.NewListOrdersService().Symbol(r.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*exchange.Order, len(orders))
	for i, o := range orders {
		res[i] = FromExtOrderToInt(o)
	}
	return res, err
}

func (c Client) ListOpenOrders(ctx context.Context, r exchange.ListOpenOrdersRequest) ([]*exchange.Order, error) {
	openOrders, err := c.b.NewListOpenOrdersService().Symbol(r.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*exchange.Order, len(openOrders))
	for i, o := range openOrders {
		res[i] = FromExtOrderToInt(o)
	}
	return res, err
}

func (c Client) GetAccount(ctx context.Context) (*exchange.Account, error) {
	acc, err := c.b.NewGetAccountService().OmitZeroBalances(true).Do(ctx)
	if err != nil {
		return nil, err
	}
	balances := make([]exchange.Balance, len(acc.Balances))
	for i, b := range acc.Balances {
		balances[i] = exchange.Balance(b)
	}
	return &exchange.Account{
		MakerCommission:  acc.MakerCommission,
		TakerCommission:  acc.TakerCommission,
		BuyerCommission:  acc.BuyerCommission,
		SellerCommission: acc.SellerCommission,
		CommissionRates:  exchange.CommissionRates(acc.CommissionRates),
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
