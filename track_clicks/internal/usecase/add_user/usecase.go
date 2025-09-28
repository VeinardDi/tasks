package add_user

import (
	"context"
	"os/user"
)

type Repo interface {
	AddUser(ctx context.Context, user *user.User) (*user.User, error)
}

type AuthClient interface {
	AuthenticateUser(ctx context.Context, user *user.User) (*user.User, error)
}
type UseCase struct {
	repo Repo
	auth AuthClient
}

func NewUseCase(repo Repo, auth AuthClient) *UseCase {
	return &UseCase{repo: repo, auth: auth}
}

func (u *UseCase) Do(ctx context.Context, user *user.User) (*user.User, error) {
	return nil, nil
}
