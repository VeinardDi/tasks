package get_counts

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type UseCase interface {
	Do(ctx context.Context, authors []string) (map[string]int64, error)
}
type Config struct {
	Timeout time.Duration
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
	g.POST("/v1/get_counts", h.getCounts)
}

type getCountsReq struct {
	Authors []string `json:"authors"`
}

type getCountsResp struct {
	Counts map[string]int64 `json:"counts"`
}

type errorResp struct {
	Error string `json:"error"`
}

func (h *Handler) getCounts(c echo.Context) error {
	var req getCountsReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResp{Error: "invalid JSON"})
	}
	if len(req.Authors) == 0 {
		return c.JSON(http.StatusBadRequest, errorResp{Error: "authors is required"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), h.cfg.Timeout)
	defer cancel()

	counts, err := h.uc.Do(ctx, req.Authors)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return c.JSON(http.StatusGatewayTimeout, errorResp{Error: "timeout"})
		default:
			return c.JSON(http.StatusBadRequest, errorResp{Error: err.Error()})
		}
	}

	return c.JSON(http.StatusOK, getCountsResp{Counts: counts})
}
