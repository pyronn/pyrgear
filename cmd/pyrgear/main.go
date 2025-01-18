package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pyrgear",
	Short: "PyRGear - A powerful tool for Python and R integration",
	Long: `PyRGear is a command-line tool that helps you seamlessly integrate Python and R workflows.
It provides various utilities to manage Python and R environments, execute scripts,
and handle data transfer between the two languages.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommands are provided, print help
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
