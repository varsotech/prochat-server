package prochat

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/community"
	"github.com/varsotech/prochat-server/internal/homeserver"
	html2 "github.com/varsotech/prochat-server/internal/homeserver/html"
	"github.com/varsotech/prochat-server/internal/imageproxy"
	"github.com/varsotech/prochat-server/internal/pkg/communitydb"
	"github.com/varsotech/prochat-server/internal/pkg/filestore"
	"github.com/varsotech/prochat-server/internal/pkg/homeserverdb"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
	"golang.org/x/sync/errgroup"
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

	homeserverDbClient, err := homeserverdb.Connect(ctx, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSL_MODE"))
	if err != nil {
		slog.Error("failed initializing homeserverdb client", "error", err, "host", os.Getenv("POSTGRES_HOST"), "user", os.Getenv("POSTGRES_USER"), "db", os.Getenv("POSTGRES_DB"), "ssl_mode", os.Getenv("POSTGRES_SSL_MODE"))
		return err
	}

	communityDbClient, err := communitydb.Connect(ctx, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_COMMUNITY_DB"), os.Getenv("POSTGRES_SSL_MODE"))
	if err != nil {
		slog.Error("failed initializing communitydb client", "error", err, "host", os.Getenv("POSTGRES_HOST"), "user", os.Getenv("POSTGRES_USER"), "db", os.Getenv("POSTGRES_DB"), "ssl_mode", os.Getenv("POSTGRES_SSL_MODE"))
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	defer redisClient.Close()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		slog.Error("failed initializing redis client", "error", err)
		return err
	}

	s3Client, err := filestore.NewS3Client(ctx, os.Getenv("AWS_S3_BUCKET"))
	if err != nil {
		slog.Error("failed initializing s3 client", "error", err)
		return err
	}

	baseFileStore := filestore.NewScope(s3Client, os.Getenv("FILE_STORE_PREFIX"))
	externalFileStore := filestore.NewScope(baseFileStore, "external")

	htmlTemplate, err := html2.NewTemplate()
	if err != nil {
		slog.Error("failed initializing html template", "error", err)
		return err
	}

	imageProxyConfig := &imageproxy.Config{
		ImageProxyBaseUrl:    os.Getenv("IMAGE_PROXY_BASE_URL"),
		ImageProxySecretKey:  os.Getenv("IMAGE_PROXY_SECRET_KEY"),
		ImageProxySecretSalt: os.Getenv("IMAGE_PROXY_SECRET_SALT"),
	}

	homeserverHost := os.Getenv("HOMESERVER_HOST")

	// HTTP routes
	homeserverRoutes := homeserver.NewRoutes(redisClient, homeserverDbClient, htmlTemplate, imageProxyConfig, homeserverHost, os.Getenv("HOMESERVER_IDENTITY_PRIVATE_KEY"), os.Getenv("HOMESERVER_IDENTITY_PUBLIC_KEY"))
	communityRoutes := community.NewRoutes(communityDbClient)
	imageProxyRoutes := imageproxy.NewRoutes(externalFileStore, imageProxyConfig)

	// Initializations
	err = community.Initialize(ctx, communityDbClient, homeserverHost)
	if err != nil {
		slog.Error("failed initializing communityserver", "error", err)
		return err
	}

	// Each routine must gracefully exit on context cancellation
	errGroup, ctx := errgroup.WithContext(ctx)

	httpServer := httputil.NewServer(ctx, os.Getenv("HTTP_SERVER_PORT"), homeserverRoutes, communityRoutes, imageProxyRoutes)
	errGroup.Go(httpServer.Serve)

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	return nil
}
