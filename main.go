package main

import (
	"os"

	"github.com/nikmd1306/cwai/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
