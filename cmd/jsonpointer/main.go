package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattn/go-jsonpointer"
	"os"
)

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s pointer\n", os.Args[0])
		os.Exit(1)
	}
	var v any
	err := json.NewDecoder(os.Stdin).Decode(&v)
	fatalIf(err)
	rv, err := jsonpointer.Get(v, os.Args[1])
	fatalIf(err)
	fmt.Println(rv)
}
