package cmd

import (
	"authex/helpers"
	"authex/model"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// used by the client to connect to the authex rest api
	restBaseURL string
	// used by the client to be non-interactive
	nonInteractive bool
	// used by the server setup to reset the database
	resetDb bool
)

func initCmd() {

	// 	DEFAULT_KEYFILE_DIRECTORY = "~/.autonity/keystore"
	// KEYFILE_DIRECTORY_ENV_VAR = "KEYFILEDIR"
	// KEYFILE_ENV_VAR = "KEYFILE"
	// KEYFILE_PASSWORD_ENV_VAR = "KEYFILEPWD"
	// WEB3_ENDPOINT_ENV_VAR = "WEB3_ENDPOINT"
	// CONTRACT_ADDRESS_ENV_VAR = "CONTRACT_ADDRESS"
	// CONTRACT_ABI_ENV_VAR = "CONTRACT_ABI"

	// CLIENT ONLY
	envRestBaseURL := helpers.EnvStr("AUTHEX_REST_URL", "http://127.0.0.1:2306")
	envNonInteractive := helpers.EnvBool("NON_INTERACTIVE", false)

	// Server and Client
	envKeystorePath := helpers.EnvStr("KEYSTORE_PATH", "./_private/keystore")
	envKeyFilePwd := helpers.EnvStr("KEYFILEPWD", "")
	envSignerAddress := helpers.EnvStr("SIGNER_ADDRESS", "")

	// Server only
	envDatabaseURI := helpers.EnvStr("DATABASE_URI", "postgres://app:app@localhost:5432/authex")
	envListenAddr := helpers.EnvStr("LISTEN_ADDR", "0.0.0.0:2306")
	envPermissioned := helpers.EnvBool("PERMISSIONED", false)
	envRpcEndpoint := helpers.EnvStr("WEB3_ENDPOINT", "https://rpc0.devnet.clearmatics.network:443/")
	envWsEndpoint := helpers.EnvStr("WEB3_WS_ENDPOINT", "wss://rpc0.devnet.clearmatics.network/ws")
	envChainID := helpers.EnvStr("CHAIN_ID", "65110000")
	envAccessControlContractAddress := helpers.EnvStr("ACCESS_CONTROL_CONTRACT", "0xCE96F4f662D807623CAB4Ce96B56A44e7cC37a48")

	// QUERY
	rootCmd.AddCommand(queryCmd)

	queryCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", envRestBaseURL, "the base URL of the REST API")

	// add the query command
	queryCmd.AddCommand(queryMarketsCmd)
	queryCmd.AddCommand(queryMarketCmd)
	queryCmd.AddCommand(queryOrderCmd)
	queryCmd.AddCommand(queryMarketQuoteCmd)
	queryCmd.AddCommand(queryMarketPriceCmd)

	// ADMIN
	rootCmd.AddCommand(adminCmd)

	adminCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", envKeystorePath, "Path to the keystore directory")
	adminCmd.PersistentFlags().StringVarP(&options.Identity.SignerAddress, "from", "f", envSignerAddress, "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	adminCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", envRestBaseURL, "the base URL of the REST API")
	adminCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", envKeyFilePwd, "the password to unlock the sender account")
	adminCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", envNonInteractive, "commands will not prompt for input (password)")

	adminCmd.AddCommand(registerMarketCmd)
	adminCmd.AddCommand(authorizeCmd)
	adminCmd.AddCommand(fundCmd)

	// ACCOUNT
	rootCmd.AddCommand(accountCmd)
	accountCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", envKeystorePath, "Path to the keystore directory")
	accountCmd.PersistentFlags().StringVarP(&options.Identity.SignerAddress, "from", "f", envSignerAddress, "the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore")
	accountCmd.PersistentFlags().StringVarP(&restBaseURL, "rest-url", "e", envRestBaseURL, "the base URL of the REST API")
	accountCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", envKeyFilePwd, "the password to unlock the sender account")
	accountCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", envNonInteractive, "commands will not prompt for input (password)")

	accountCmd.AddCommand(bidLimitCmd)
	accountCmd.AddCommand(bidMarketCmd)
	accountCmd.AddCommand(askLimitCmd)
	accountCmd.AddCommand(askMarketCmd)
	accountCmd.AddCommand(cancelOrderCmd)
	accountCmd.AddCommand(withdrawCmd)

	// SERVER
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&options.DB.URI, "database-uri", "d", envDatabaseURI, "Database URI")
	serverCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", envKeystorePath, "Path to the keystore directory")
	serverCmd.PersistentFlags().StringVarP(&options.Identity.SignerAddress, "signer-address", "f", envSignerAddress, "the address to send the transaction from (must be an account in the keystore)")
	serverCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", envKeyFilePwd, "Password for key file (or use env var 'KEYFILEPWD')")

	serverCmd.PersistentFlags().StringVarP(&options.Web.ListenAddr, "listen-address", "l", envListenAddr, "Address the REST server listen to (format host:port)")
	serverCmd.PersistentFlags().BoolVar(&options.Web.Permissioned, "permissioned", envPermissioned, "when the flag is set only authorized accounts are allowed to interact authex")
	serverCmd.PersistentFlags().StringVarP(&options.Network.RPCEndpoint, "rpc-endpoint", "r", envRpcEndpoint, "RPC endpoint (defaults to WEB3_ENDPOINT env var if set)")
	serverCmd.PersistentFlags().StringVarP(&options.Network.WSEndpoint, "ws-endpoint", "w", envWsEndpoint, "WS endpoint (defaults to WEB3_WS_ENDPOINT env var if set)")
	serverCmd.PersistentFlags().StringVarP(&options.Network.ChainID, "chain-id", "I", envChainID, "The chain ID of the network to connect to")

	serverCmd.PersistentFlags().StringVarP(&options.Identity.AccessContractAddress, "access-control-contract", "z", envAccessControlContractAddress, "The contract address to look for access control (must be an AcccessControl contract)")

	setupCmd.Flags().BoolVar(&resetDb, "reset", false, "Reset the database before setup")

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

func requireFromAddress(cmd *cobra.Command, args []string) {
	// check that the from address is set
	if helpers.IsEmpty(options.Identity.SignerAddress) {
		fmt.Println("error: signer address is not set, use the --from flag to set the address")
		os.Exit(1)
	}
}
