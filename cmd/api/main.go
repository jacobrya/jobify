package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abzalserikbay/jobify/internal/config"
	"github.com/abzalserikbay/jobify/internal/handler"
	"github.com/abzalserikbay/jobify/internal/middleware"
	postgresrepo "github.com/abzalserikbay/jobify/internal/repository/postgres"
	rediscache "github.com/abzalserikbay/jobify/internal/repository/redis"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/internal/worker"
	"github.com/abzalserikbay/jobify/pkg/hasher"
	jwtpkg "github.com/abzalserikbay/jobify/pkg/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Database
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logger.Error("postgres ping failed", "err", err)
		os.Exit(1)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()

	// Repos
	userRepo := postgresrepo.NewUserRepo(db)
	profileRepo := postgresrepo.NewProfileRepo(db)
	jobRepo := postgresrepo.NewJobRepo(db)
	appRepo := postgresrepo.NewApplicationRepo(db)
	jobCache := rediscache.NewJobCache(rdb)
	rlStore := rediscache.NewRateLimitStore(rdb)

	// Services
	h := hasher.New()
	jwt := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTExpiry)

	authSvc := service.NewAuthService(userRepo, profileRepo, h, jwt)
	userSvc := service.NewUserService(userRepo, profileRepo)
	jobSvc := service.NewJobService(jobRepo, jobCache)
	appSvc := service.NewApplicationService(appRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	jobHandler := handler.NewJobHandler(jobSvc, userSvc)
	appHandler := handler.NewApplicationHandler(appSvc)

	// Router
	router := handler.NewRouter(&handler.Deps{
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		JobHandler:         jobHandler,
		ApplicationHandler: appHandler,
		JWT:                jwt,
		RateLimitStore:     rlStore,
		RateLimitPerMin:    cfg.RateLimitPerMin,
	})

	// Apply logger middleware
	loggedRouter := middleware.Logger(logger)(router)

	// Background worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aggregator := worker.NewJobAggregator(jobRepo, logger, 6*time.Hour, cfg.RemotiveAPIURL)
	go aggregator.Start(ctx)

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      loggedRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	logger.Info("server started", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
