package cmd

import (
	"authex/helpers"
	"fmt"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Group of query commands",
}

var queryMarketsCmd = &cobra.Command{
	Use:     "markets",
	Aliases: []string{"get-markets"},
	Short:   "Get all markets",
	Example: `authex account markets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return queryMarkets(restBaseURL)
	},
}

func queryMarkets(url string) error {
	// send the request
	code, data, err := helpers.Get(fmt.Sprint(url, "/markets"))
	if err != nil {
		println("error getting markets:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}

var queryMarketCmd = &cobra.Command{
	Use:     "market <market-address>",
	Aliases: []string{"get-market"},
	Short:   "Query a market",
	Args:    cobra.ExactArgs(1),
	Example: `authex query market 0x123...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return queryMarket(restBaseURL, args[0])
	},
}

func queryMarket(url, address string) error {
	// send the request
	code, data, err := helpers.Get(fmt.Sprint(url, "/markets/", address))
	if err != nil {
		println("error getting markets:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}

var queryOrderCmd = &cobra.Command{
	Use:     "order <order-id>",
	Aliases: []string{"get-order"},
	Short:   "Query an order",
	Args:    cobra.ExactArgs(1),
	Example: `authex query order abcd-adf-123... `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return queryOrder(restBaseURL, args[0])
	},
}

func queryOrder(url, address string) error {
	// send the request
	code, data, err := helpers.Get(fmt.Sprint(url, "/orders/", address))
	if err != nil {
		println("error getting orders:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}
