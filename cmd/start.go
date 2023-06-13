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

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the CLOB server",
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
