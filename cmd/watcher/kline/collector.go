package kline

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"

	"crypto_bot/pkg/exchange/binance"
	"crypto_bot/pkg/storage/pgdb"
	"crypto_bot/pkg/watcher/kline"
)

var (
	CollectorFlags = struct {
		ConnStr   string
		Symbol    string
		Interval  string
		ChunkSize int
		Debug     bool
	}{}
	CollectCmd = &cobra.Command{
		Use:   "collect",
		Short: "Save new klines from exchange to db",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			signal.Notify(interrupt, os.Kill)
			go func() { <-interrupt; log.Println("keyboard interruption"); cancel() }()

			c := binance.NewClient("", "")

			conn, err := pgx.Connect(ctx, CollectorFlags.ConnStr)
			if err != nil {
				return err
			}
			db := pgdb.NewClient(conn)

			log.Printf("Starting watcher %s interval=%s chunk-size=%d",
				CollectorFlags.Symbol, CollectorFlags.Interval, CollectorFlags.ChunkSize)
			watcher := kline.
				NewWatcher(c, db).
				SetSymbols(CollectorFlags.Symbol).
				SetInterval(CollectorFlags.Interval).
				SetChunkSize(CollectorFlags.ChunkSize).
				SetErrorHandler(func(err error) {
					log.Printf("ERROR: %s", err)
				}).
				SetDebug(CollectorFlags.Debug)

			if err = watcher.Start(ctx); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	flags := CollectCmd.Flags()
	flags.StringVar(&CollectorFlags.ConnStr, "conn-str", "", "connection string")
	flags.StringVar(&CollectorFlags.Symbol, "symbol", "BTCUSDT", "symbol to watch")
	flags.StringVar(&CollectorFlags.Interval, "interval", "1m", "interval of symbol to watch")
	flags.IntVar(&CollectorFlags.ChunkSize, "chunk-size", 50, "chunk size to write to db")
	flags.BoolVarP(&CollectorFlags.Debug, "debug", "v", false, "chunk size to write to db")

	if err := cobra.MarkFlagRequired(flags, "conn-str"); err != nil {
		log.Fatal(err)
	}
}
