package cmd

import (
	"authex/helpers"
	"authex/model"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------
// user Commands
// -----------------------------------------------------------------------------

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Group of user commands",
}

var withdrawCmd = &cobra.Command{
	Use: "withdraw",
}

func order(URL string, from string, market string, symbol string, side string, size string, price string) error {
	sizeUint, err := strconv.ParseUint(size, 10, 64)
	if err != nil {
		return err
	}
	o := model.Order{
		Market: market,
		Side:   side,
		Size:   uint(sizeUint),
		Price:  price,
	}
	// sign the message
	signature, err := helpers.Sign(options.Identity.KeystorePath, from, o)
	if err != nil {
		fmt.Println("error signing the message:", err)
		return err
	}
	r := &model.SignedRequest[model.Order]{
		Signature: signature,
		Payload:   o,
	}
	// send the request
	code, data, err := helpers.Post(fmt.Sprintf("%s/orders", URL), r)
	if err != nil {
		fmt.Println("error creating order:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}

var askLimitCmd = &cobra.Command{
	Use:     "ask <market-address> <asset> <size> <price>",
	Aliases: []string{"ask-limit", "sell-limit", "sell"},
	Short:   "Submit a new order",
	Args:    cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], model.SideAsk, args[2], args[3])
	},
}

var bidLimitCmd = &cobra.Command{
	Use:     "bid <market-address> <asset> <size> <price>",
	Aliases: []string{"bid-limit", "sell", "sell-limit"},
	Short:   "Submit a new buy limit order",
	Args:    cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], model.SideBid, args[2], args[3])
	},
}

var askMarketCmd = &cobra.Command{
	Use:     "ask <market-address> <asset> <size>",
	Aliases: []string{"ask-limit", "sell-limit", "sell"},
	Short:   "Submit a new order",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], model.SideAsk, args[2], "")
	},
}

var bidMarketCmd = &cobra.Command{
	Use:     "bid <market-address> <asset> <size>",
	Aliases: []string{"bid-limit", "sell", "sell-limit"},
	Short:   "Submit a new buy limit order",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], model.SideBid, args[2], "")
	},
}

var cancelCmd = &cobra.Command{
	Use:   "cancel-order <order-id>",
	Short: "Cancel an order",
}
