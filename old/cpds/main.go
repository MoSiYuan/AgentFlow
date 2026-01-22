package main

import (
	"os"

	"github.com/jiangxiaolong/cpds-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
