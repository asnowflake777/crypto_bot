package exchange

type Client interface {
	KlinesService
	WsKlinesService
	OrderService
	AccountService
}
