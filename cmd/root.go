package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rea",
	Short: "rea is a document renderer",
	Long: `A document renderer that makes your document generation easy.
Code is hosted at https://github.com/microfast-ch/rea/`,
	//	Run: func(cmd *cobra.Command, args []string) {
	//		// Do Stuff Here
	//	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
