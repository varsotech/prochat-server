package main

import (
	"log/slog"
	"os"

	"github.com/varsotech/prochat-server/internal/prochat"
)

func main() {
	err := prochat.Run()
	if err != nil {
		slog.Error("error running server", "error", err)
		os.Exit(1)
	}
}
