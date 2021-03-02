package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"log"
)

const defaultDocDir = "./"

var docCmd = &cobra.Command{
	Use:   "doc [path/to/file.md]",
	Short: "Generates markdown documentation",
	Long:  `Generates markdown documentation.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if args[0] == "" {
			dir = defaultDocDir
		} else {
			dir = args[0]
		}

		err := doc.GenMarkdownTree(rootCmd, dir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}
