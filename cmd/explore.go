package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"petezalew.ski/zip/zipfile"
)

var exploreCmd = &cobra.Command{
	Use: "explore",
	Short: "Display the structure and metadata of zip files",
	Long: "Print the local headers, central directory, and end record of one or more zip files.",
	Run: runExplore,
}

func runExplore(cmd *cobra.Command, args []string) {
	for _, filename := range args {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer f.Close()
		zf, err := zipfile.Parse(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, header := range zf.LocalHeaders {
			fmt.Printf("%+v\n", header)
			fmt.Println(header.GetContent())
		}
	}

}
