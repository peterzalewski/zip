package main

import (
	"flag"
	"fmt"
	"os"

	"petezalew.ski/zip/zipfile"
)

func main() {
	flag.Parse()

	for _, filename := range flag.Args() {
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
