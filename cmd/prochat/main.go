package main

import (
	"github.com/varsotech/prochat-server/internal/prochat"
	"log/slog"
	"os"
)

func main() {
	err := prochat.Run()
	if err != nil {
		slog.Error("prochat: error running server", "error", err)
		os.Exit(1)
	}
}
