package track_click

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Config struct {
	Timeout time.Duration
}

type UseCase interface {
	Do(ctx context.Context, authorID string, userID string) error
}

type Handler struct {
	uc  UseCase
	cfg Config
}

func New(uc UseCase, cfg Config) *Handler {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}

	return &Handler{uc: uc, cfg: cfg}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/v1/click", h.Handle)
}

type clickReq struct {
	AuthorID string `json:"author_id"`
	UserID   string `json:"user_id"`
}

type errorResp struct {
	Error string `json:"error"`
}

func (h *Handler) Handle(c echo.Context) error {
	var req clickReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResp{Error: "invalid JSON"})
	}
	if req.AuthorID == "" || req.UserID == "" {
		return c.JSON(http.StatusBadRequest, errorResp{Error: "author_id and user_id are required"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), h.cfg.Timeout)
	defer cancel()

	if err := h.uc.Do(ctx, req.AuthorID, req.UserID); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return c.JSON(http.StatusGatewayTimeout, errorResp{Error: "timeout"})
		default:
			return c.JSON(http.StatusInternalServerError, errorResp{Error: "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}
