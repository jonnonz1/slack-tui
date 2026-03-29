package main

import (
	"fmt"
	"os"

	"github.com/jonnonz1/slack-tui/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
