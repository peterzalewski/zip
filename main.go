package main

import (
	"flag"
	"fmt"
	"os"

	"petezalew.ski/zip/structure"
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
		headers, err := structure.Parse(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, header := range headers {
			fmt.Printf("%+v\n", header)
		}
	}
}
