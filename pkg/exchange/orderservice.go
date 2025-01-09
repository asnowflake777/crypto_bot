package exchange

import "context"

type OrderService interface {
	CreateOrder(context.Context, CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(context.Context, ReadOrderRequest) (*Order, error)
	CancelOrder(context.Context, CancelOrderRequest) (*CancelOrderResponse, error)
	ListOrders(context.Context, ListOrdersRequest) ([]*Order, error)
	ListOpenOrders(context.Context, ListOpenOrdersRequest) ([]*Order, error)
}

type Order struct {
	Symbol                   string          `json:"symbol"`
	OrderID                  int64           `json:"orderId"`
	OrderListId              int64           `json:"orderListId"`
	ClientOrderID            string          `json:"clientOrderId"`
	Price                    float64         `json:"price"`
	OrigQuantity             float64         `json:"origQty"`
	ExecutedQuantity         float64         `json:"executedQty"`
	CummulativeQuoteQuantity float64         `json:"cummulativeQuoteQty"`
	Status                   OrderStatusType `json:"status"`
	TimeInForce              TimeInForceType `json:"timeInForce"`
	Type                     OrderType       `json:"type"`
	Side                     SideType        `json:"side"`
	StopPrice                float64         `json:"stopPrice"`
	IcebergQuantity          float64         `json:"icebergQty"`
	Time                     int64           `json:"time"`
	UpdateTime               int64           `json:"updateTime"`
	IsWorking                bool            `json:"isWorking"`
	IsIsolated               bool            `json:"isIsolated"`
	OrigQuoteOrderQuantity   float64         `json:"origQuoteOrderQty"`
}

type CancelOrderResponse struct {
	Symbol                   string                  `json:"symbol"`
	OrigClientOrderID        string                  `json:"origClientOrderId"`
	OrderID                  int64                   `json:"orderId"`
	OrderListID              int64                   `json:"orderListId"`
	ClientOrderID            string                  `json:"clientOrderId"`
	TransactTime             int64                   `json:"transactTime"`
	Price                    float64                 `json:"price"`
	OrigQuantity             float64                 `json:"origQty"`
	ExecutedQuantity         float64                 `json:"executedQty"`
	CummulativeQuoteQuantity float64                 `json:"cummulativeQuoteQty"`
	Status                   OrderStatusType         `json:"status"`
	TimeInForce              TimeInForceType         `json:"timeInForce"`
	Type                     OrderType               `json:"type"`
	Side                     SideType                `json:"side"`
	SelfTradePreventionMode  SelfTradePreventionMode `json:"selfTradePreventionMode"`
}

type CreateOrderResponse struct {
	Symbol                   string
	OrderID                  int64
	ClientOrderID            string
	TransactTime             int64
	Price                    float64
	OrigQuantity             float64
	ExecutedQuantity         float64
	CummulativeQuoteQuantity float64
	IsIsolated               bool
	Status                   OrderStatusType
	TimeInForce              TimeInForceType
	Type                     OrderType
	Side                     SideType
	Fills                    []*Fill
	MarginBuyBorrowAmount    float64
	MarginBuyBorrowAsset     float64
	SelfTradePreventionMode  SelfTradePreventionMode
}

type Fill struct {
	TradeID         int64
	Price           float64
	Quantity        float64
	Commission      float64
	CommissionAsset float64
}

type (
	SideType                string
	OrderType               string
	TimeInForceType         string
	SelfTradePreventionMode string
	OrderStatusType         string
)

const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
	OrderTypeStopLoss        OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"

	TimeInForceTypeGTC TimeInForceType = "GTC"
	TimeInForceTypeIOC TimeInForceType = "IOC"
	TimeInForceTypeFOK TimeInForceType = "FOK"

	SelfTradePreventionModeNone        SelfTradePreventionMode = "NONE"
	SelfTradePreventionModeExpireTaker SelfTradePreventionMode = "EXPIRE_TAKER"
	SelfTradePreventionModeExpireBoth  SelfTradePreventionMode = "EXPIRE_BOTH"
	SelfTradePreventionModeExpireMaker SelfTradePreventionMode = "EXPIRE_MAKER"

	OrderStatusTypeNew             OrderStatusType = "NEW"
	OrderStatusTypePartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusTypeFilled          OrderStatusType = "FILLED"
	OrderStatusTypeCanceled        OrderStatusType = "CANCELED"
	OrderStatusTypePendingCancel   OrderStatusType = "PENDING_CANCEL"
	OrderStatusTypeRejected        OrderStatusType = "REJECTED"
	OrderStatusTypeExpired         OrderStatusType = "EXPIRED"
	OrderStatusExpiredInMatch      OrderStatusType = "EXPIRED_IN_MATCH" // STP Expired
)

type CreateOrderRequest struct {
	Symbol      string
	Quantity    float64
	Price       float64
	Side        SideType
	Type        OrderType
	InTimeForce TimeInForceType
}

type ReadOrderRequest struct {
	ID     int64
	Symbol string
}

type ListOrdersRequest struct {
	Symbol string
}

type ListOpenOrdersRequest struct {
	Symbol string
}

type CancelOrderRequest struct {
	ID     int64
	Symbol string
}
