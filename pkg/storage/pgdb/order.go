package pgdb

import "context"

type Order struct {
	ID       int64
	Symbol   string
	Price    float64
	Quantity float64
	Type     string
	Side     string
}

type CreateOrderRequest struct {
	Symbol   string
	Price    float64
	Quantity float64
	Type     string
	Side     string
}

func (c *Client) CreateOrder(ctx context.Context, r CreateOrderRequest) (*Order, error) {
	return nil, nil
}

type ReadOrderRequest struct {
	ID int64
}

func (c *Client) ReadOrder(ctx context.Context, r ReadKlinesRequest) (*Order, error) {
	return nil, nil
}

type ReadOrdersRequest struct {
	Symbol string
}

func (c *Client) ReadOrders(ctx context.Context, r ReadKlinesRequest) ([]*Order, error) {
	return nil, nil
}

type UpdateOrderRequest struct {
	ID       int64
	Symbol   string
	Price    float64
	Quantity float64
	Type     string
	Side     string
}

func (c *Client) UpdateOrder(ctx context.Context, r UpdateOrderRequest) (*Order, error) {
	return nil, nil
}

func (c *Client) DeleteOrder(ctx context.Context, id int64) error {
	return nil
}
