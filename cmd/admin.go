package cmd

import (
	"authex/helpers"
	"authex/model"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {

}

// -----------------------------------------------------------------------------
// admin Commands
// -----------------------------------------------------------------------------

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Group of admin commands",
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

		m := model.Market{
			BaseSymbol:  base[0],
			QuoteSymbol: quote[0],
		}
		if len(base) > 1 {
			m.BaseAddress = base[1]
		}
		if len(quote) > 1 {
			m.QuoteAddress = quote[1]
		}
		// sign the message
		signature, err := helpers.Sign(options.Identity.KeystorePath, from, m)
		if err != nil {
			fmt.Println("error signing the message:", err)
			return err
		}

		mr := &model.SignedRequest[model.Market]{
			Signature: signature,
			Payload:   m,
		}

		// send the request
		code, data, err := helpers.Post(fmt.Sprintf("%s/markets", restBaseURL), mr)
		if err != nil {
			fmt.Println("error creating market:", err)
			return err
		}
		fmt.Println("response code:", code)
		fmt.Println("response body:", data)
		// open the database connection
		return
	}
}

// registerMarketCmd represents the registerMarket command
var authorizeCmd = &cobra.Command{
	Use:   "authorize [account_address]",
	Short: "Authorize a new account to trade",
	Args:  cobra.ExactArgs(1),
	Long: `A market is a pair of tokens or assets that can be traded together.

	For example, the market ETH/USDT is the pair of tokens ETH and USDT.

	Markets are identified by a base token and a quote token.
	The base token is the token that is being bought or sold, and the quote token is the token that is used to pay for the base token.
	`,
	Example: `authex admin authorize 0x1234...`,
	RunE:    registerMarket(options),
}

var fundCmd = &cobra.Command{
	Use:     "fund <account-address> <asset-address> <amount>",
	Short:   "Fund an account with an asset (modify the account balance in AutHEx)",
	Args:    cobra.ExactArgs(3),
	Example: `authex admin fund 0x1234... 0x1234... 1000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fund(restBaseURL, from, args[0], args[1], args[2])
	},
}

func fund(url, signer, account, asset, amount string) error {
	f := model.Funding{
		Account: account,
		Asset:   asset,
		Amount:  amount,
	}
	// sign the message
	signature, err := helpers.Sign(options.Identity.KeystorePath, signer, f)
	if err != nil {
		fmt.Println("error signing the message:", err)
		return err
	}
	r := &model.SignedRequest[model.Funding]{
		Signature: signature,
		Payload:   f,
	}
	fmt.Printf("sending request: %#v", r)
	// send the request
	code, data, err := helpers.Post(fmt.Sprintf("%s/fund", restBaseURL), r)
	if err != nil {
		fmt.Println("error funding account:", err)
		return err
	}
	helpers.PrintResponse(code, data)
	return nil
}
