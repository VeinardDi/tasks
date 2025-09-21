package get_counts

import (
	"context"
	"errors"
)

var (
	manyAuthorsErr = errors.New("too many authors")
)

type Store interface {
	CountYesterday(ctx context.Context, authorIDs []string) map[string]int64
}

type Config struct {
	MaxAuthorsPerQuery int
}

type UseCase struct {
	store Store
	cfg   Config
}

func New(store Store, cfg Config) *UseCase {
	if cfg.MaxAuthorsPerQuery <= 0 {
		cfg.MaxAuthorsPerQuery = 1000
	}
	return &UseCase{store: store, cfg: cfg}
}

func (uc *UseCase) Do(ctx context.Context, authorIDs []string) (map[string]int64, error) {
	if len(authorIDs) > uc.cfg.MaxAuthorsPerQuery {
		return nil, manyAuthorsErr
	}
	return uc.store.CountYesterday(ctx, authorIDs), nil
}
