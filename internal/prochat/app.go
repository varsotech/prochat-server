package prochat

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html"
	"github.com/varsotech/prochat-server/internal/pkg/httpserver"
	"github.com/varsotech/prochat-server/internal/pkg/postgres"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
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

	authRoutes := authhttp.NewRoutes(postgresClient, redisClient)
	htmlRoutes, err := html.NewRoutes()
	if err != nil {
		slog.Error("error initializing html routes", "error", err)
		return err
	}

	httpServer := httpserver.New(ctx, os.Getenv("HTTP_SERVER_PORT"), authRoutes, htmlRoutes)

	// Each routine must gracefully exit on context cancellation
	errGroup.Go(httpServer.Serve)

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	return nil
}
