package cmd

import (
	"authex/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	restBaseURL = "http://127.0.0.1:2306"
	from        string // the address to send the transaction from
)

func initCmd() {
	// QUERY
	rootCmd.AddCommand(queryCmd)

	queryCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", restBaseURL, "the base URL of the REST API")

	// add the query command
	queryCmd.AddCommand(queryMarketsCmd)
	queryCmd.AddCommand(queryMarketCmd)
	queryCmd.AddCommand(queryOrderCmd)
	queryCmd.AddCommand(queryMarketQuoteCmd)

	// ADMIN
	rootCmd.AddCommand(adminCmd)

	adminCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")
	adminCmd.PersistentFlags().StringVar(&from, "from", "", "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	adminCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", restBaseURL, "the base URL of the REST API")

	adminCmd.AddCommand(registerMarketCmd)
	adminCmd.AddCommand(authorizeCmd)
	adminCmd.AddCommand(fundCmd)

	// ACCOUNT
	rootCmd.AddCommand(accountCmd)
	accountCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")
	accountCmd.PersistentFlags().StringVar(&from, "from", "", "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	accountCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", restBaseURL, "the base URL of the REST API")

	accountCmd.AddCommand(bidLimitCmd)
	accountCmd.AddCommand(bidMarketCmd)
	accountCmd.AddCommand(askLimitCmd)
	accountCmd.AddCommand(askMarketCmd)
	accountCmd.AddCommand(cancelCmd)
	accountCmd.AddCommand(withdrawCmd)

	// SERVER
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&options.DB.URI, "database-uri", "d", "postgres://app:app@localhost:5432/authex", "Database URI")
	viper.BindPFlag("DB.URI", serverCmd.PersistentFlags().Lookup("database-uri"))
	serverCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")
	serverCmd.PersistentFlags().StringVarP(&options.Identity.KeyFile, "keyfile", "f", "_private/UTC--2023-03-26T09-46-49.997099000Z--e2fb069045dfb19f3dd2b95a5a09d6f62984932d", "Encrypted private key file to import")
	serverCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", "puravida", "Password for key file (or use env var 'KEYFILEPWD')")

	serverCmd.PersistentFlags().StringVarP(&options.Web.ListenAddr, "listen-address", "l", "0.0.0.0:2306", "Address the REST server listen to (format host:port)")
	serverCmd.PersistentFlags().BoolVar(&options.Web.Permissioned, "permissioned", false, "when the flag is set only authorized accounts are allowed to interact authex")
	serverCmd.PersistentFlags().StringVarP(&options.Network.RPCEndpoint, "rpc-endpoint", "r", "https://rpc0.devnet.clearmatics.network:443/", "RPC endpoint (defaults to WEB3_ENDPOINT env var if set)")
	serverCmd.PersistentFlags().StringVarP(&options.Network.WSEndpoint, "ws-endpoint", "w", "wss://rpc0.devnet.clearmatics.network/ws", "WS endpoint (defaults to WEB3_WS_ENDPOINT env var if set)")
	serverCmd.PersistentFlags().StringVarP(&options.Network.ChainID, "chain-id", "I", "65110000", "The chain ID of the network to connect to")

	serverCmd.PersistentFlags().StringVarP(&options.Identity.AccessContractAddress, "access-control-contract", "z", "0xCE96F4f662D807623CAB4Ce96B56A44e7cC37a48", "The contract address to look for access control (must be an AcccessControl contract)")

	serverCmd.AddCommand(setupCmd)
	serverCmd.AddCommand(startCmd)
}

// options hold the settings for the server
var options = &model.Settings{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "authex",
	Short: "Ethereum Exchange Server",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v string) error {
	options.Version = v
	initCmd()
	rootCmd.Version = options.Version
	return rootCmd.Execute()
}
