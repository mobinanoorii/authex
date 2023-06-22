package cmd

import (
	"authex/helpers"
	"authex/model"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------
// admin Commands
// -----------------------------------------------------------------------------

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Group of admin commands",
}

// registerMarketCmd represents the registerMarket command.
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return registerMarket(restBaseURL, args[0], args[1])
	},
}

func registerMarket(url, baseStr, quoteStr string) (err error) {
	// base token
	base := strings.Split(baseStr, ":")
	// quote token
	quote := strings.Split(quoteStr, ":")

	market := model.Market{
		BaseSymbol:  base[0],
		QuoteSymbol: quote[0],
	}
	if len(base) > 1 {
		market.BaseAddress = base[1]
	}
	if len(quote) > 1 {
		market.QuoteAddress = quote[1]
	}
	// sign the message

	signature, err := helpers.Sign(
		options.Identity.KeystorePath,
		options.Identity.SignerAddress,
		options.Identity.Password,
		!nonInteractive,
		market,
	)
	if err != nil {
		println("error signing the message:", err)
		return err
	}

	mr := &model.SignedRequest[model.Market]{
		Signature: signature,
		Payload:   market,
	}

	// send the request
	code, data, err := helpers.Post(fmt.Sprint(url, "/admin/markets"), mr)
	if err != nil {
		println("error creating market:", err)
		return err
	}
	println("response code:", code)
	println("response body:", data)
	// open the database connection
	return
}

// registerMarketCmd represents the registerMarket command.
var authorizeCmd = &cobra.Command{
	Use:     "authorize [account_address]",
	Short:   "Authorize a new account to trade",
	Args:    cobra.ExactArgs(1),
	Example: `authex admin authorize 0x1234...`,
}

var fundCmd = &cobra.Command{
	Use:     "fund <account-address> <asset-address> <amount>",
	Short:   "Fund an account with an asset (modify the account balance in AutHEx)",
	Args:    cobra.ExactArgs(3),
	Example: `authex admin fund 0x1234... 0x1234... 1000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fund(restBaseURL, args[0], args[1], args[2])
	},
}

func fund(url, account, asset, amount string) error {
	funding := model.Funding{
		Account: account,
		Asset:   asset,
		Amount:  amount,
	}
	// sign the message
	signature, err := helpers.Sign(
		options.Identity.KeystorePath,
		options.Identity.SignerAddress,
		options.Identity.Password,
		!nonInteractive,
		funding,
	)
	if err != nil {
		println("error signing the message:", err)
		return err
	}
	r := &model.SignedRequest[model.Funding]{
		Signature: signature,
		Payload:   funding,
	}
	// send the request
	code, data, err := helpers.Post(fmt.Sprint(url, "/admin/accounts/fund"), r)
	if err != nil {
		println("error funding account:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}
