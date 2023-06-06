package cmd

import (
	"authex/db"
	"authex/model"
	"authex/network"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the AutHEx server",
	RunE:  setup(options),
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
