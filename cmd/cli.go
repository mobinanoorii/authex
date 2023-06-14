package cmd

import (
	"authex/model"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	restBaseURL string
	from        string // the address to send the transaction from
)

func init() {

	defaultRestBaseurl := "http://127.0.0.1:2306"

	adminCmd.PersistentFlags().StringVar(&from, "from", "", "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	adminCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", defaultRestBaseurl, "the base URL of the REST API")

	adminCmd.AddCommand(registerMarketCmd)
	rootCmd.AddCommand(adminCmd)
	// user commands

	accountCmd.PersistentFlags().StringVar(&from, "from", "", "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	accountCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", defaultRestBaseurl, "the base URL of the REST API")

	accountCmd.AddCommand(withdrawCmd)
	rootCmd.AddCommand(accountCmd)

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
		signature, err := sign(from, m)
		if err != nil {
			fmt.Println("error signing the message:", err)
			return err
		}

		mr := &model.SignedRequest[model.Market]{
			Signature: signature,
			Payload:   m,
		}

		// send the request
		code, data, err := post(fmt.Sprintf("%s/markets", restBaseURL), mr)
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
	Use:     "fund [account_address] [asset_address] [amount]",
	Short:   "Fund an account with an off-chain asset",
	Args:    cobra.ExactArgs(3),
	Example: `authex admin authorize 0x1234...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fund(restBaseURL, from, args[0], args[1], args[2])
	},
}

func fund(url, signer, account, asset, amount string) error {
	return nil
	
}

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

var askLimitCmd = &cobra.Command{
	Use:     "ask [market-id] [asset] [size] [price]",
	Aliases: []string{"ask-limit", "sell-limit", "sell"},
	Short:   "Submit a new order",
}

var bidLimitCmd = &cobra.Command{
	Use:     "bid [market-id] [asset] [size] [price]",
	Aliases: []string{"bid-limit", "sell", "sell-limit"},
	Short:   "Submit a new buy limit order",
}

var askMarketCmd = &cobra.Command{
	Use:     "ask [market-id] [asset] [size]",
	Aliases: []string{"ask-limit", "sell-limit", "sell"},
	Short:   "Submit a new order",
}

var bidMarketCmd = &cobra.Command{
	Use:     "bid [market-id] [asset] [size]",
	Aliases: []string{"bid-limit", "sell", "sell-limit"},
	Short:   "Submit a new buy limit order",
}

var cancelCmd = &cobra.Command{
	Use:   "cancel-order [order-id]",
	Short: "Cancel an order",
}

// -----------------------------------------------------------------------------
// utils
// -----------------------------------------------------------------------------

func sign(address string, msg any) (signature string, err error) {
	ks := keystore.NewKeyStore(options.Identity.KeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	var signer accounts.Account
	// get the account
	if len(ks.Accounts()) == 0 {
		err = fmt.Errorf("no accounts found in the keystore")
		return
	}
	if len(ks.Accounts()) == 1 {
		signer = ks.Accounts()[0]
	} else {
		for _, acc := range ks.Accounts() {
			if acc.Address.String() == address {
				signer = acc
				break
			}
		}
		if signer.Address.String() == "" {
			err = fmt.Errorf("no account found for address %s", address)
			return
		}
	}
	// unlock the account
	pass := passwordPrompt(signer.Address.String())
	if err = ks.Unlock(signer, pass); err != nil {
		fmt.Println("error unlocking account:", err)
		return
	}
	// prepare the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	msgBytes = crypto.Keccak256(msgBytes)
	// sign the message
	sigBytes, err := ks.SignHash(signer, msgBytes)
	if err != nil {
		return
	}
	signature = hex.EncodeToString(sigBytes)
	return
}

// Post make a json rest request
func post(url string, data interface{}) (code int, body map[string]any, err error) {
	client := &http.Client{
		Timeout: time.Hour * 2,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	code = resp.StatusCode
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&body)
	return
}

// PasswordPrompt asks for a string value using the label.
// The entered value will not be displayed on the screen
// while typing.
func passwordPrompt(account string) string {
	label := fmt.Sprintf("Enter password for account %s:", account)
	var s string
	for {
		fmt.Fprint(os.Stderr, label+" ")
		b, _ := term.ReadPassword(int(syscall.Stdin))
		s = string(b)
		if s != "" {
			break
		}
	}
	fmt.Println()
	return s
}
