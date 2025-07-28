package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/matheusrb95/fibergraph/internal/api"
	"github.com/matheusrb95/fibergraph/internal/aws"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	_ = godotenv.Load()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return fmt.Errorf("load aws config. %w", err)
	}

	db, err := database.OpenDB()
	if err != nil {
		return fmt.Errorf("open db. %w", err)
	}

	services := aws.NewServices(cfg)
	err = services.SNS.Ping()
	if err != nil {
		return fmt.Errorf("sns client not working. %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	models := data.NewModels(db)
	srv := api.NewServer(logger, models, services)

	httpServer := &http.Server{
		Addr:    ":4000",
		Handler: srv,
	}
	go func() {
		logger.Info("starting server", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err.Error())
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error(err.Error())
		}
	}()

	wg.Wait()

	return nil
}
