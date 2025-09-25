package prochat

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html"
	"github.com/varsotech/prochat-server/internal/oauth"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
	"github.com/varsotech/prochat-server/internal/pkg/postgres"
	"github.com/varsotech/prochat-server/internal/pkg/s3util"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
)

func Run() error {
	_ = godotenv.Load()

	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	}

	slog.Info("initializing server")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	postgresClient, err := postgres.Connect(ctx, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSL_MODE"))
	if err != nil {
		slog.Error("failed initializing postgres client", "error", err, "host", os.Getenv("POSTGRES_HOST"), "user", os.Getenv("POSTGRES_USER"), "db", os.Getenv("POSTGRES_DB"), "ssl_mode", os.Getenv("POSTGRES_SSL_MODE"))
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		slog.Error("failed initializing redis client", "error", err)
		return err
	}

	s3Client, err := s3util.NewClient(ctx, os.Getenv("AWS_S3_BUCKET"))
	if err != nil {
		slog.Error("failed initializing s3 client", "error", err)
		return err
	}

	htmlTemplate, err := html.NewTemplate()
	if err != nil {
		slog.Error("failed initializing html template", "error", err)
		return err
	}

	authenticator := authhttp.NewAuthenticator(redisClient)
	authHttpService := authhttp.New(postgresClient, redisClient, authenticator)
	htmlRoutes := html.NewRoutes(htmlTemplate, authenticator)
	oauthHttpRoutes := oauth.NewRoutes(redisClient, s3Client, authenticator, httputil.NewClient(), htmlTemplate)
	httpServer := httputil.NewServer(ctx, os.Getenv("HTTP_SERVER_PORT"), authHttpService, oauthHttpRoutes, htmlRoutes)

	// Each routine must gracefully exit on context cancellation
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(httpServer.Serve)

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	return nil
}
