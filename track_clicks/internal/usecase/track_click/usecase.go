package track_click

import (
	"context"
)

type Store interface {
	AddClick(ctx context.Context, authorID, userID string) error
}

type UseCase struct {
	store Store
}

func New(store Store) *UseCase {
	return &UseCase{store: store}
}

func (uc *UseCase) Do(ctx context.Context, authorID, userID string) error {
	return uc.store.AddClick(ctx, authorID, userID)
}
