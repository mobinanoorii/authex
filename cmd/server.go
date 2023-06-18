package cmd

import (
	"authex/clob"
	"authex/db"
	"authex/model"
	"authex/network"
	"authex/web"
	"fmt"

	"github.com/spf13/cobra"
)

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
		// restore markets
		markets, err := db.GetMarkets()
		if err != nil {
			err = fmt.Errorf("error getting the markets: %v", err)
			return
		}
		for _, market := range markets {
			clob.OpenMarket(market.Address)
		}
		// TODO: restore orders

		// start the network client
		nodeCli, err := network.NewNodeClient(options, db.Transfers)
		if err != nil {
			err = fmt.Errorf("error setting up the node client: %v", err)
			return
		}
		go nodeCli.Run()
		// get the token list and send them to the node client
		tokens, err := db.GetAssetAddressesByClass(model.AssetERC20)
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
			fmt.Println("error connecting to the database")
			return err
		}
		err = network.Setup(options)
		if err != nil {
			fmt.Println("error in network setup")
			return err
		}
		return nil
	}
}
