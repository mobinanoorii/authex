/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"authex/db"
	"authex/model"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(registerMarketCmd)
}

// registerMarketCmd represents the registerMarket command
var registerMarketCmd = &cobra.Command{
	Use:   "register-market",
	Short: "Register a new market",
	Args:  cobra.ExactArgs(2),
	Long: `A market is a pair of tokens or assets that can be traded together.

	For example, the market ETH/USDT is the pair of tokens ETH and USDT.

	Markets are identified by a base token and a quote token.
	The base token is the token that is being bought or sold, and the quote token is the token that is used to pay for the base token.
	`,
	Example: `authex register-market BASET QUOTET:0x1234...`,
	RunE:    registerMarket(options),
}

func registerMarket(options *model.Settings) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, args []string) (err error) {
		// base token
		base := strings.Split(args[0], ":")
		// quote token
		quote := strings.Split(args[1], ":")

		// TODO: check if the address is correct
		// // start the network client
		// nodeCli, err := network.NewNodeClient(options)
		// if err != nil {
		// 	err = fmt.Errorf("error setting up the node client: %v", err)
		// 	return
		// }

		m := &model.Market{
			BaseSymbol:  base[0],
			QuoteSymbol: quote[0],
		}
		if len(base) > 1 {
			m.BaseAddress = base[1]
		}
		if len(quote) > 1 {
			m.QuoteAddress = quote[1]
		}
		// open the database connection
		db, err := db.NewConnection(options.DB.URI)
		if err != nil {
			err = fmt.Errorf("error connecting to the database: %v", err)
			return
		}
		if err = db.SaveMarket(m); err != nil {
			err = fmt.Errorf("error saving market: %v", err)
			return
		}
		return
	}
}
