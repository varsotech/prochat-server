package prochat

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/varsotech/prochat-server/internal/database/postgres"
)

func Run() error {
	slog.Info("initializing server")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_ = godotenv.Load()

	_, err := postgres.Connect(ctx, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSL_MODE"))
	if err != nil {
		return err
	}

	slog.Info("server is ready to accept connections")

	return nil
}
