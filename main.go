package main

import (
	"log/slog"
	"os"
	"sql2pb/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
}
