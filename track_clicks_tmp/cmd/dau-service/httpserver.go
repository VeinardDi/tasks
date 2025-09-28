package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"example.com/dau/internal/date"
	"example.com/dau/internal/dau"
	"example.com/dau/internal/repository"
)

type HTTPServer struct {
	service      *dau.Service
	shuttingDown int32 // атомарный флаг завершения сервера
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				c.AbortWithStatus(http.StatusRequestTimeout)
				return
			}
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		default:
		}

		c.Next()
	}
}

func (s *HTTPServer) GracefulShutdownMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if atomic.LoadInt32(&s.shuttingDown) == 1 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Server is shutting down",
				"code":  "SHUTTING_DOWN",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *HTTPServer) SetShuttingDown() {
	atomic.StoreInt32(&s.shuttingDown, 1)
}

func (s *HTTPServer) Event(c *gin.Context) {
	var request dau.EventRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	select {
	case <-c.Request.Context().Done():
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service is shutting down"})
		return
	default:
	}

	if err := s.service.Event(c.Request.Context(), &request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

func (s *HTTPServer) Dau(c *gin.Context) {
	rawList := c.QueryArray("authors_list")
	var authorsList []int
	for _, rawID := range rawList {
		id, err := strconv.Atoi(rawID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID format"})
			return
		}

		authorsList = append(authorsList, id)
	}

	select {
	case <-c.Request.Context().Done():
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service is shutting down"})
		return
	default:
	}

	result, err := s.service.Dau(c.Request.Context(), authorsList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ServerInfo содержит HTTP сервер и HTTPServer для управления
type ServerInfo struct {
	Server     *http.Server
	HTTPServer *HTTPServer
}

func NewServer(port int) *ServerInfo {
	dateService := date.NewService()
	repo := repository.New(dateService)
	service := dau.NewService(dateService, repo)
	httpServer := &HTTPServer{service: service}

	gin.SetMode(gin.ReleaseMode)
	handler := gin.Default()

	handler.Use(TimeoutMiddleware(30 * time.Second))
	handler.Use(httpServer.GracefulShutdownMiddleware())

	handler.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Основные endpoints
	handler.POST("/event", httpServer.Event)
	handler.GET("/dau", httpServer.Dau)

	return &ServerInfo{
		Server: &http.Server{
			Handler:           handler,
			Addr:              fmt.Sprintf(":%d", port),
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		},
		HTTPServer: httpServer,
	}
}
