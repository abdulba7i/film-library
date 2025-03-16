package main

import (
	"film-library/internal/config"
	"film-library/internal/lib/logger/handlers/slogpretty"
	"film-library/internal/storage/postgres"
	"fmt"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting film-library", slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	storage, err := postgres.Connect(cfg.Database)
	if err != nil {
		// log.Error("failed to init storage", "error: %v", sl.Err(err))
		log.Error("failed to init storage", "error", err)
		fmt.Println("Storage connection error:", err)
		os.Exit(1)
	}

	if err != nil {
		log.Error("failed to change actor", slog.String("error", err.Error()))
		os.Exit(1)
	}

	_ = storage
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		// log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		log = setupPrettySlog() // здесь преукрасили вывод логов для удобства
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)

	default:
		panic("not supported env")
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
