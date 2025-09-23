package prochat

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/httpserver"
	"golang.org/x/sync/errgroup"
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

	postgresClient, err := postgres.Connect(ctx, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSL_MODE"))
	if err != nil {
		slog.Error("error initializing postgres client", "error", err, "host", os.Getenv("POSTGRES_HOST"), "user", os.Getenv("POSTGRES_USER"), "db", os.Getenv("POSTGRES_DB"), "ssl_mode", os.Getenv("POSTGRES_SSL_MODE"))
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		slog.Error("error initializing redis client", "error", err)
		return err
	}

	errGroup, ctx := errgroup.WithContext(ctx)

	// Each routine must gracefully exit on context cancellation
	errGroup.Go(httpserver.New(ctx, os.Getenv("HTTP_SERVER_PORT"), postgresClient).Serve)

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	return nil
}
