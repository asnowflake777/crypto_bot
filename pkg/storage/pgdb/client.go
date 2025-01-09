package pgdb

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
)

type Client struct {
	conn *pgx.Conn
}

func NewClient(conn *pgx.Conn) *Client {
	return &Client{conn: conn}
}

func createKlineTableIfNotExists(ctx context.Context, tx pgx.Tx, symbol, interval string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("create table if not exists kline_%s_%s as table kline with no data;",
		strings.ToLower(symbol), strings.ToLower(interval)))
	return err
}
