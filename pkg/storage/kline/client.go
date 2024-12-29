package kline

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"log"
	"strings"
)

type Client struct {
	conn *pgx.Conn
}

func NewClient(conn *pgx.Conn) *Client {
	return &Client{conn: conn}
}

type Kline struct {
	OpenTime  int64
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	CloseTime int64
	TradeNum  int64
}

type ReadKlineRequest struct {
	Symbol   string
	Interval string
	OpenTime int64
}

func (c *Client) ReadKline(ctx context.Context, req *ReadKlineRequest) (*Kline, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := sq.
		Select("open_time", "open", "high", "low", "close", "volume", "close_time", "trade_num").
		From(fmt.Sprintf("kline_%s_%s", strings.ToLower(req.Symbol), strings.ToLower(req.Interval))).
		Where("open_time = ?", req.OpenTime).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	kline := Kline{}
	err = tx.
		QueryRow(ctx, query, args...).
		Scan(&kline.OpenTime, &kline.Open, &kline.High, &kline.Low, &kline.Close,
			&kline.Volume, &kline.CloseTime, &kline.TradeNum)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

type ReadKlinesRequest struct {
	Symbol   string
	Interval string
	OpenTime int64
	Limit    uint64
	Offset   uint64
}

func (c *Client) ReadKlines(ctx context.Context, req *ReadKlinesRequest) ([]*Kline, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := sq.
		Select("open_time", "open", "high", "low", "close", "volume", "close_time", "trade_num").
		From(fmt.Sprintf("kline_%s_%s", strings.ToLower(req.Symbol), strings.ToLower(req.Interval))).
		Where("open_time = ?", req.OpenTime).
		Limit(req.Limit).
		Offset(req.Offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	klines := make([]*Kline, 0, 100)
	for rows.Next() {
		kline := &Kline{}
		err = rows.Scan(&kline.OpenTime, &kline.Open, &kline.High, &kline.Low, &kline.Close,
			&kline.Volume, &kline.CloseTime, &kline.TradeNum)
		if err != nil {
			return nil, err
		}
		klines = append(klines, kline)
	}
	return klines, rows.Err()
}

type WriteKlineRequest struct {
	Symbol   string
	Interval string
	Kline    *Kline
}

func (c *Client) WriteKline(ctx context.Context, req *WriteKlineRequest) (*Kline, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("Rollback failed: %v", err)
		}
	}()
	if err = createTableIfNotExists(ctx, tx, req.Symbol, req.Interval); err != nil {
		return nil, err
	}

	query, args, err := sq.
		Insert(fmt.Sprintf("kline_%s_%s", strings.ToLower(req.Symbol), strings.ToLower(req.Interval))).
		Columns("open_time", "open", "high", "low", "close", "volume", "close_time", "trade_num").
		Values(req.Kline.OpenTime, req.Kline.Open, req.Kline.High, req.Kline.Low, req.Kline.Close,
			req.Kline.Volume, req.Kline.CloseTime, req.Kline.TradeNum).
		Suffix("ON CONFLICT DO NOTHING").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return req.Kline, nil
}

type WriteKlinesRequest struct {
	Symbol   string
	Interval string
	Klines   []*Kline
}

func (c *Client) WriteKlines(ctx context.Context, req *WriteKlinesRequest) ([]*Kline, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("Rollback failed: %v", err)
		}
	}()
	if err = createTableIfNotExists(ctx, tx, req.Symbol, req.Interval); err != nil {
		return nil, err
	}

	query := sq.
		Insert(fmt.Sprintf("kline_%s_%s", strings.ToLower(req.Symbol), strings.ToLower(req.Interval))).
		Columns("open_time", "open", "high", "low", "close", "volume", "close_time", "trade_num").
		Suffix("ON CONFLICT DO NOTHING").
		PlaceholderFormat(sq.Dollar)

	for _, kline := range req.Klines {
		query = query.Values(
			kline.OpenTime, kline.Open, kline.High, kline.Low, kline.Close,
			kline.Volume, kline.CloseTime, kline.TradeNum)
	}
	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, queryStr, args...); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return req.Klines, nil
}

func createTableIfNotExists(ctx context.Context, tx pgx.Tx, symbol, interval string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("create table if not exists kline_%s_%s as table kline with no data;",
		strings.ToLower(symbol), strings.ToLower(interval)))
	return err
}