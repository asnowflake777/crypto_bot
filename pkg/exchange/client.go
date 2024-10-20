package exchange

type Client interface {
	KlinesService
	WsKlinesService
}
