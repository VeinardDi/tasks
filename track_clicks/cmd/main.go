package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	get_counts_handler "track_clicks/handler/get_counts"
	track_click_handler "track_clicks/handler/track_click"
	mid "track_clicks/internal/middleware"
	"track_clicks/internal/pkg"
	"track_clicks/internal/usecase/get_counts"
	"track_clicks/internal/usecase/track_click"
)

func main() {
	// TODO вынести в конфиг
	// Настраиваем slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	store := pkg.NewStore(pkg.SystemClock{})

	ucTrack := track_click.New(store)

	ucGetCounts := get_counts.New(store, get_counts.Config{MaxAuthorsPerQuery: 1000})

	handlerGetCounts := get_counts_handler.New(ucGetCounts,
		get_counts_handler.Config{Timeout: 2 * time.Second})
	handlerTrackClick := track_click_handler.New(ucTrack,
		track_click_handler.Config{Timeout: 2 * time.Second})

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetOutput(os.Stdout)

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(mid.SlogLogger())

	// пробы
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/ready", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	api := e.Group("/api")

	handlerGetCounts.Register(api)
	handlerTrackClick.Register(api)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      e,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := e.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server start failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	slog.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		if err := e.Close(); err != nil {
			slog.Error("force close failed", "err", err)
		}
	}
	slog.Info("server stopped")
}
