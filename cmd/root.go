package cmd

import (
	"authex/model"

	"github.com/spf13/cobra"
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

	serverCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")

	rootCmd.Version = options.Version
	return rootCmd.Execute()
}
