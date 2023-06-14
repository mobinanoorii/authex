package cmd

import (
	"authex/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*

//   -r, --rpc-endpoint URL  RPC endpoint (defaults to WEB3_ENDPOINT env var if set)

To make transfers
-I, --chain-id INTEGER          integer representing EIP155 chainId.
-p, --password TEXT  Password for key file (or use env var 'KEYFILEPWD')
-k, --keyfile PATH   Encrypted private key file

To get the privileges
-z --authorization-contract the contract address ot look for authorization

Database
-d, --database-uri postgres://username:password@localhost:5432/database_name

/home/andrea/.autonity/keystore/UTC--2023-03-26T09-46-49.997099000Z--e2fb069045dfb19f3dd2b95a5a09d6f62984932d

*/

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

	rootCmd.PersistentFlags().StringVarP(&options.DB.URI, "database-uri", "d", "postgres://app:app@localhost:5432/authex", "Database URI")
	viper.BindPFlag("DB.URI", rootCmd.PersistentFlags().Lookup("database-uri"))
	rootCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")
	rootCmd.PersistentFlags().StringVarP(&options.Identity.KeyFile, "keyfile", "f", "_private/UTC--2023-03-26T09-46-49.997099000Z--e2fb069045dfb19f3dd2b95a5a09d6f62984932d", "Encrypted private key file to import")
	rootCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", "puravida", "Password for key file (or use env var 'KEYFILEPWD')")

	rootCmd.PersistentFlags().StringVarP(&options.Web.ListenAddr, "listen-address", "l", "0.0.0.0:2306", "Address the REST server listen to (format host:port)")
	rootCmd.PersistentFlags().BoolVar(&options.Web.Permissioned, "permissioned", false, "when the flag is set only authorized accounts are allowed to interact authex")
	rootCmd.PersistentFlags().StringVarP(&options.Network.RPCEndpoint, "rpc-endpoint", "r", "https://rpc0.devnet.clearmatics.network:443/", "RPC endpoint (defaults to WEB3_ENDPOINT env var if set)")
	rootCmd.PersistentFlags().StringVarP(&options.Network.WSEndpoint, "ws-endpoint", "w", "wss://rpc0.devnet.clearmatics.network/ws", "WS endpoint (defaults to WEB3_WS_ENDPOINT env var if set)")
	rootCmd.PersistentFlags().StringVarP(&options.Network.ChainID, "chain-id", "I", "65110000", "The chain ID of the network to connect to")

	rootCmd.PersistentFlags().StringVarP(&options.Identity.AccessContractAddress, "access-control-contract", "z", "0xCE96F4f662D807623CAB4Ce96B56A44e7cC37a48", "The contract address to look for access control (must be an AcccessControl contract)")

	rootCmd.Version = options.Version
	return rootCmd.Execute()
}
