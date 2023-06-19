package cmd

import (
	"authex/helpers"
	"authex/model"
	"errors"
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
	Use:     "withdraw",
	Short:   `Withdraw tokens from the exchange.`,
	Example: `authex account withdraw <asset-address> <amount>`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return withdraw(restBaseURL, args[0], args[1])
	},
}

func withdraw(url string, asset string, amount string) error {
	return fmt.Errorf("not implemented yet")
}

func order(url string, from string, market string, size string, price string, side string) error {
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
	signature, err := helpers.Sign(options.Identity.KeystorePath, from, options.Identity.Password, !nonInteractive, o)
	if err != nil {
		err = errors.Join(errors.New("error signing the message"), err)
		return err
	}
	r := &model.SignedRequest[model.Order]{
		Signature: signature,
		Payload:   o,
	}
	// send the request
	code, data, err := helpers.Post(fmt.Sprint(url, "/account/orders"), r)
	if err != nil {
		err = errors.Join(errors.New("error creating order"), err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}

func cancelOrder(url string, id string) error {
	o := model.Order{
		ID:   id,
		Side: model.CancelOrder,
	}
	// sign the message
	signature, err := helpers.Sign(options.Identity.KeystorePath, from, options.Identity.Password, !nonInteractive, o)
	if err != nil {
		err = errors.Join(errors.New("error signing the message"), err)
		return err
	}
	r := &model.SignedRequest[model.Order]{
		Signature: signature,
		Payload:   o,
	}
	// send the request
	code, data, err := helpers.Post(fmt.Sprint(url, "/account/orders/cancel"), r)
	if err != nil {
		err = errors.Join(errors.New("error cancelling order"), err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}

var askLimitCmd = &cobra.Command{
	Use:     "ask <market-address> <size> <price>",
	Aliases: []string{"ask-limit", "sell-limit", "sell", "offer"},
	Short:   "Submit a new order",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], args[2], model.SideAsk)
	},
}

var bidLimitCmd = &cobra.Command{
	Use:     "bid <market-address> <size> <price>",
	Aliases: []string{"bid-limit", "buy-limit", "buy"},
	Short:   "Submit a new buy limit order",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], args[2], model.SideBid)
	},
}

var askMarketCmd = &cobra.Command{
	Use:     "ask-market <market-address> <size>",
	Aliases: []string{"ask-market", "sell-market", "offer-market"},
	Short:   "Submit a new market order",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], "", model.SideAsk)
	},
}

var bidMarketCmd = &cobra.Command{
	Use:     "bid-market <market-address> <size>",
	Aliases: []string{"bid-market", "buy-market"},
	Short:   "Submit a new buy limit order",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return order(restBaseURL, from, args[0], args[1], "", model.SideBid)
	},
}

var cancelOrderCmd = &cobra.Command{
	Use:     "cancel-order <order-id>",
	Short:   "Cancel an order",
	Aliases: []string{"cancel"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cancelOrder(restBaseURL, args[0])
	},
}
