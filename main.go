/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"authex/clob"
	"authex/db"
	"authex/model"
	"authex/network"
	"authex/web"

	"github.com/labstack/gommon/log"
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

var (
	// Version is the version of the application
	Version = "dev"
)

func main() {
	var options = &model.Settings{Version: Version}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "authex",
		Short: "Ethereum Exchange Server",
	}

	rootCmd.PersistentFlags().StringVarP(&options.DB.URI, "database-uri", "d", "postgres://app:app@localhost:5432/authex", "Database URI")
	viper.BindPFlag("DB.URI", rootCmd.PersistentFlags().Lookup("database-uri"))
	rootCmd.PersistentFlags().StringVarP(&options.Identity.KeystorePath, "keystore-path", "k", "./_private/keystore", "Path to the keystore directory")
	rootCmd.PersistentFlags().StringVarP(&options.Identity.KeyFile, "keyfile", "f", "_private/UTC--2023-03-26T09-46-49.997099000Z--e2fb069045dfb19f3dd2b95a5a09d6f62984932d", "Encrypted private key file to import")
	rootCmd.PersistentFlags().StringVarP(&options.Identity.Password, "password", "p", "puravida", "Password for key file (or use env var 'KEYFILEPWD')")

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the CLOB server",
		RunE:  start(options),
	}
	startCmd.Flags().StringVarP(&options.Web.ListenAddr, "listen-address", "l", "0.0.0.0:2306", "Address the REST server listen to (format host:port)")
	startCmd.Flags().StringVarP(&options.Network.RPCEndpoint, "rpc-endpoint", "r", "https://rpc0.devnet.clearmatics.network:443/", "RPC endpoint (defaults to WEB3_ENDPOINT env var if set)")
	startCmd.Flags().StringVarP(&options.Network.WSEndpoint, "ws-endpoint", "w", "wss://ws0.devnet.clearmatics.network:443/", "WS endpoint (defaults to WEB3_WS_ENDPOINT env var if set)")
	startCmd.Flags().StringVarP(&options.Network.ChainID, "chain-id", "I", "65110000", "The chain ID of the network to connect to")

	startCmd.Flags().StringVarP(&options.Identity.AccessContractAddress, "access-control-contract", "z", "0xCE96F4f662D807623CAB4Ce96B56A44e7cC37a48", "The contract address to look for access control (must be an AcccessControl contract)")
	startCmd.Flags().StringVarP(&options.Identity.TokenAddress, "token-contract", "t", "0x9096065c6ed910e2e3db3e0a442aa0ec975557bb", "The contract address to listen for transfer events (must be an ERC20 contract)")

	rootCmd.AddCommand(startCmd)

	var setupComd = &cobra.Command{
		Use:   "setup",
		Short: "Setup the CLOB server",
		RunE:  setup(options),
	}
	rootCmd.AddCommand(setupComd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// runFunction create the new resolver driver from the options and start the server.
func start(options *model.Settings) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		// open the database connection
		db, err := db.NewConnection(options.DB.URI)
		if err != nil {
			log.Fatalf("error connecting to the database: %v", err)
		}
		go db.Run()

		// start the clob engine
		clob := clob.NewPool(db.Matches)
		go clob.Run()

		// start the network client
		nodeCli, err := network.NewNodeClient(options)
		if err != nil {
			log.Fatalf("error setting up the node client: %v", err)
		}
		nodeCli.Run()
		// finally start the server
		resolverDriver, err := web.NewAuthexServer(options, clob.Inbound, nodeCli)
		if err != nil {
			return err
		}
		return resolverDriver.Start()
	}
}

func setup(options *model.Settings) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		// open the database connection
		err := db.Setup(options.DB.URI)
		if err != nil {
			log.Fatalf("error connecting to the database: %v", err)
		}
		err = network.Setup(options)
		if err != nil {
			log.Fatalf("error in network setup: %v", err)
		}
		return nil
	}
}
