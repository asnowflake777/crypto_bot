package pgdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Client struct {
	conn *pgx.Conn
}

func NewClient(conn *pgx.Conn) *Client {
	return &Client{conn: conn}
}

func createKlineTableIfNotExists(ctx context.Context, tx pgx.Tx, symbol, interval string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("create table if not exists kline_%s_%s (like kline including all);",
		strings.ToLower(symbol), strings.ToLower(interval)))
	return err
}
