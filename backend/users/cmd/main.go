package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cwby333/user-microservice/internal/adapters/repository/postgres"
	"github.com/Cwby333/user-microservice/internal/adapters/tokenStorage/redis"
	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/server"
	userrouter "github.com/Cwby333/user-microservice/internal/adapters/transport/http/userRouter"
	"github.com/Cwby333/user-microservice/internal/config"
	"github.com/Cwby333/user-microservice/internal/migrations"
	userservice "github.com/Cwby333/user-microservice/internal/service/userService"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 3)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, os.Kill)

	logger := initLogger()

	go func() {
		sig := <-signalChan

		logger.Error("received signal", slog.String("signal", sig.String()))
		cancel()
	}()

	cfg := config.MustLoad()

	pgCfg := postgres.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DB:       cfg.DB.DB,
		MaxConns: cfg.DB.MaxConns,
		MinConns: cfg.DB.MinConns,
	}

	pg, err := postgres.New(ctx, pgCfg)
	if err != nil {
		logger.Error("postgres connect", slog.String("error", err.Error()))
		return
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DB)

	err = migrations.Migrate(connStr)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	cfgRedis := redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	redis, err := redis.New(ctx, cfgRedis)
	if err != nil {
		slog.Info("redis connect", slog.String("error", err.Error()))
		return
	}

	cfgJWT := userservice.JWTConfig{
		SecretKey:      cfg.JWT.SecretKey,
		Issuer:         cfg.JWT.Issuer,
		AccessExpired:  cfg.JWT.JWTAccess.Expired,
		RefreshExpired: cfg.JWT.JWTRefresh.Expired,
	}
	userService := userservice.New(pg, pg, redis, redis, cfgJWT)

	userRouter := userrouter.New(userService, userService, logger)
	userRouter.Run()

	cfgServer := server.Config{
		Address:         cfg.Server.Address,
		IdleTimeout:     cfg.Server.IdleTimeout,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
	}

	serv := server.New(cfgServer, userRouter.Mux)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("server start", slog.String("address", cfg.Server.Address))
		return serv.Server.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("start shutdown")

		ctxShutdown, canc := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer canc()

		return serv.Server.Shutdown(ctxShutdown)
	})

	if err := g.Wait(); err != nil {
		logger.Info("server stopped", slog.String("error", err.Error()))
	}
}

func initLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	return logger
}
