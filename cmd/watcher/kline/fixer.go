package kline

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"

	"crypto_bot/pkg/exchange/binance"
	"crypto_bot/pkg/helpers/gapfixer"
	"crypto_bot/pkg/storage/pgdb"
)

const timeLayout = "2006-01-02_15:04:05"

var (
	FixGapsFlags = struct {
		ConnStr   string
		Symbol    string
		Interval  string
		From      string
		To        string
		ChunkSize int
	}{}

	FixGapsCmd = &cobra.Command{
		Use:   "fix-gaps",
		Short: "Fix gaps in klines",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, err := time.Parse(timeLayout, FixGapsFlags.From)
			if err != nil {
				return fmt.Errorf("parse from: %s", err)
			}

			to := time.Now()
			if FixGapsFlags.To != "" {
				if to, err = time.Parse(timeLayout, FixGapsFlags.To); err != nil {
					return fmt.Errorf("parse to: %s", err)
				}
			}

			ctx, cancel := context.WithCancel(context.Background())

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			signal.Notify(interrupt, os.Kill)
			go func() { <-interrupt; log.Println("keyboard interruption"); cancel() }()

			c := binance.NewClient("", "")

			conn, err := pgx.Connect(ctx, FixGapsFlags.ConnStr)
			if err != nil {
				return err
			}
			db := pgdb.NewClient(conn)

			err = gapfixer.FixGaps(ctx, c, db, FixGapsFlags.Symbol, FixGapsFlags.Interval, from, to, FixGapsFlags.ChunkSize)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	flags := FixGapsCmd.Flags()
	flags.StringVar(&FixGapsFlags.ConnStr, "conn-str", "", "pg db connection string")
	flags.StringVar(&FixGapsFlags.Symbol, "symbol", "BTCUSDT", "symbol to fix gaps for")
	flags.StringVar(&FixGapsFlags.Interval, "interval", "1m", "interval to fix gaps")
	flags.StringVar(&FixGapsFlags.From, "from", "2017-08-17_4:00:00", "to fix gaps from date")
	flags.StringVar(&FixGapsFlags.To, "to", "", "to fix gaps to date")
	flags.IntVar(&FixGapsFlags.ChunkSize, "chunk-size", 100, "chunk size")
}
