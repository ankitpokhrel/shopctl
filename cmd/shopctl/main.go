package main

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/shopctl/internal/cmd/root"
)

func main() {
	rootCmd := root.NewCmdRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
