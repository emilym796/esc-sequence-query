package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	var r io.Reader
	if len(os.Args) > 1 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "escq:", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	} else {
		r = os.Stdin
	}

	data, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "escq:", err)
		os.Exit(1)
	}

	if err := Scan(data, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
