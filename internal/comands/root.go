package comands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	RootCmd.AddCommand(RenameCmd)
	RootCmd.AddCommand(ExifCmd)
}
