/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"

	"authex/cmd"
)

// Version hold the version of the program
var Version = "dev"

func main() {
	if err := cmd.Execute(Version); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
