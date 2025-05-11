package kline

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "kline",
	Short: "Commands for watching for exchange's klines",
}

func init() {
	RootCmd.AddCommand(CollectCmd)
	RootCmd.AddCommand(FixGapsCmd)
}
