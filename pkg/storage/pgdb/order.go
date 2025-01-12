package pgdb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

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
	UserUID  int64
}

func (c *Client) CreateOrder(ctx context.Context, r CreateOrderRequest) (*Order, error) {
	queryStr, args, err := sq.
		Insert("orders").
		Columns("symbol", "price", "quantity", "type", "side", "user_uid").
		Values(r.Symbol, r.Price, r.Quantity, r.Type, r.Side, r.UserUID).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var id int64
	if err = c.conn.QueryRow(ctx, queryStr, args...).Scan(&id); err != nil {
		return nil, err
	}
	return &Order{
		ID:       id,
		Symbol:   r.Symbol,
		Price:    r.Price,
		Quantity: r.Quantity,
		Type:     r.Type,
		Side:     r.Side,
	}, nil
}

type ReadOrderRequest struct {
	ID int64
}

func (c *Client) ReadOrder(ctx context.Context, r ReadOrderRequest) (*Order, error) {
	queryStr, args, err := sq.
		Select("id", "symbol", "price", "quantity", "type", "side").
		From("orders").
		Where(sq.Eq{"id": r.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var o Order
	if err = c.conn.QueryRow(ctx, queryStr, args...).Scan(&o.ID, &o.Symbol, &o.Price, &o.Quantity, &o.Type, &o.Side); err != nil {
		return nil, err
	}
	return &o, nil
}

type ReadOrdersRequest struct {
	UserUID int64
	Symbol  string
}

func (c *Client) ReadOrders(ctx context.Context, r ReadOrdersRequest) ([]*Order, error) {
	query := sq.
		Select("id", "symbol", "price", "quantity", "type", "side").
		From("orders").
		Where(sq.Eq{"user_uid": r.UserUID}).
		PlaceholderFormat(sq.Dollar)
	if r.Symbol != "" {
		query = query.Where(sq.Eq{"symbol": r.Symbol})
	}
	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := c.conn.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var os []*Order
	for rows.Next() {
		var o Order
		if err = rows.Scan(&o.ID, &o.Symbol, &o.Price, &o.Quantity, &o.Type, &o.Side); err != nil {
			return nil, err
		}
		if rows.Err() != nil {
			return nil, err
		}
		os = append(os, &o)
	}
	return os, nil
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
	queryStr, args, err := sq.
		Update("orders").
		Set("symbol", r.Symbol).
		Set("price", r.Price).
		Set("quantity", r.Quantity).
		Set("type", r.Type).
		Set("side", r.Side).
		Where(sq.Eq{"id": r.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return nil, err
	}
	return &Order{
		ID:       r.ID,
		Symbol:   r.Symbol,
		Price:    r.Price,
		Quantity: r.Quantity,
		Type:     r.Type,
		Side:     r.Side,
	}, nil
}

type DeleteOrderRequest struct {
	ID int64
}

func (c *Client) DeleteOrder(ctx context.Context, r DeleteOrderRequest) error {
	queryStr, args, err := sq.Delete("orders").Where(sq.Eq{"id": r.ID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return err
	}
	return nil
}
