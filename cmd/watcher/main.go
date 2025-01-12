package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5"

	"crypto_bot/pkg/exchange"
	"crypto_bot/pkg/exchange/binance"
	"crypto_bot/pkg/storage/pgdb"
)

var flags = struct {
	connStr   string
	symbol    string
	interval  string
	chunkSize int
}{}

func init() {
	flag.StringVar(&flags.connStr, "conn-str", "", "connection string")
	flag.StringVar(&flags.symbol, "symbol", "", "symbol to watch")
	flag.StringVar(&flags.interval, "interval", "", "interval of symbol to watch")
	flag.IntVar(&flags.chunkSize, "chunk-size", 50, "chunk size to write to db")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, os.Kill)
	go func() { <-interrupt; log.Println("keyboard interruption"); cancel() }()

	c := binance.NewClient("", "")

	conn, err := pgx.Connect(ctx, flags.connStr)
	if err != nil {
		log.Fatal(err)
	}
	db := pgdb.NewClient(conn)

	events, errs, err := c.WsKlines(ctx, exchange.WsKlineRequest{Symbol: flags.symbol, Interval: flags.interval})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for err := range errs {
			log.Printf("ERROR: %s", err)
		}
	}()

	if err = watch(ctx, events, db); err != nil {
		log.Fatal(err)
	}
}

func watch(ctx context.Context, events <-chan *exchange.WsKlineEvent, db *pgdb.Client) error {
	klinesToWrite := make([]*pgdb.Kline, 0, flags.chunkSize)
	for event := range events {
		if !event.Kline.IsFinal {
			continue
		}
		klinesToWrite = append(klinesToWrite, &pgdb.Kline{
			OpenTime:  event.Kline.StartTime,
			Open:      event.Kline.Open,
			High:      event.Kline.High,
			Low:       event.Kline.Low,
			Close:     event.Kline.Close,
			Volume:    event.Kline.Volume,
			CloseTime: event.Kline.EndTime,
			TradeNum:  event.Kline.TradeNum,
		})
		if len(klinesToWrite) >= flags.chunkSize {
			_, err := db.WriteKlines(ctx, pgdb.WriteKlinesRequest{
				Symbol:   flags.symbol,
				Interval: flags.interval,
				Klines:   klinesToWrite,
			})
			if err != nil {
				return err
			}
			klinesToWrite = klinesToWrite[:0]
		}
	}
	return nil
}
