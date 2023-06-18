/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"authex/cmd"
)

// Version hold the version of the program
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	if err := cmd.Execute(version); err != nil {
		println(err)
		os.Exit(1)
	}
}
