package cmd

import (
	"authex/clob"
	"authex/db"
	"authex/model"
	"authex/network"
	"authex/web"
	"log"

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
		authex, err := web.NewAuthexServer(options, clob.Inbound, nodeCli)
		if err != nil {
			return err
		}
		return authex.Start()
	}
}
