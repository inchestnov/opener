package main

import (
	"fmt"
	"os"

	"github.com/inchestnov/opener/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "opener: %v\n", err)
		os.Exit(1)
	}
}
