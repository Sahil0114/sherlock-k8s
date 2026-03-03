package main

import (
	"os"

	"github.com/kube-sherlock/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
