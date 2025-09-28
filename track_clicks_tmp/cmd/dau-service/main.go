package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	port            int
	shutdownTimeout time.Duration
)

func init() {
	flag.IntVar(&port, "port", 8080, "server port")
	flag.DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "graceful shutdown timeout")
	flag.Parse()
}

func main() {
	log.Printf("Starting DAU service on port %d", port)

	serverInfo := NewServer(port)

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем сервер в горутине
	serverDone := make(chan error, 1)
	go func() {
		log.Printf("Server starting on :%d", port)
		if err := serverInfo.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverDone <- err
		}
		close(serverDone)
	}()

	// Ждем сигнал завершения
	<-ctx.Done()
	log.Printf("Received shutdown signal, gracefully shutting down...")

	// Устанавливаем флаг завершения сервера
	serverInfo.HTTPServer.SetShuttingDown()
	log.Printf("Shutdown flag set, rejecting new requests")

	// Создаем контекст с таймаутом для shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Пытаемся корректно завершить сервер
	if err := serverInfo.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Printf("Server gracefully stopped")
	}

	// Ждем завершения сервера или таймаута
	select {
	case err := <-serverDone:
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	case <-shutdownCtx.Done():
		log.Printf("Shutdown timeout exceeded, forcing exit")
	}

	log.Printf("DAU service stopped")
}
