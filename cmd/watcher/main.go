package main

import (
	"log"

	"github.com/spf13/cobra"

	"crypto_bot/cmd/watcher/kline"
)

var RootCmd = &cobra.Command{
	Use:   "watcher",
	Short: "Programs for watch exchange",
}

func init() {
	RootCmd.AddCommand(kline.RootCmd)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
