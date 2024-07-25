package main

import (
	"github.com/mooncake9527/sql2pb/cmd"
	"log/slog"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
}
