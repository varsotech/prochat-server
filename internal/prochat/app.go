package prochat

import (
	"context"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/varsotech/prochat-server/internal/database/postgres"
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_ = godotenv.Load()

	_, err := postgres.Connect(ctx)
	if err != nil {
		return err
	}

	return nil
}
