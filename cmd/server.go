package cmd

import (
	"authex/clob"
	"authex/db"
	"authex/model"
	"authex/network"
	"authex/web"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	serverCmd.PersistentFlags().StringVarP(&options.DB.URI, "database-uri", "d", "postgres://app:app@localhost:5432/authex", "Database URI")
	viper.BindPFlag("DB.URI", serverCmd.PersistentFlags().Lookup("database-uri"))
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
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Group of server commands",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the AutHEx server",
	RunE:  start(options),
}

// runFunction create the new resolver driver from the options and start the server.
func start(options *model.Settings) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) (err error) {
		// open the database connection
		db, err := db.NewConnection(options)
		if err != nil {
			err = fmt.Errorf("error connecting to the database: %v", err)
			return
		}
		go db.Run()

		// start the clob engine
		clob := clob.NewPool(db.Matches)
		go clob.Run()

		// start the network client
		nodeCli, err := network.NewNodeClient(options, db.Transfers)
		if err != nil {
			err = fmt.Errorf("error setting up the node client: %v", err)
			return
		}
		go nodeCli.Run()
		// get the token list and send them to the node client
		tokens, err := db.GetTokenAddresses()
		if err != nil {
			err = fmt.Errorf("error getting the token list: %v", err)
			return
		}
		for _, token := range tokens {
			nodeCli.Tokens <- token
		}

		// finally start the server
		authex, err := web.NewAuthexServer(options, clob, nodeCli, db)
		if err != nil {
			err = fmt.Errorf("error starting the server: %v", err)
			return
		}
		return authex.Start()
	}
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the AutHEx server",
	RunE:  setup(options),
}

func setup(options *model.Settings) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		// open the database connection
		err := db.Setup(options)
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
