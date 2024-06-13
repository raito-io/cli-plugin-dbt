package raito

import (
	"context"
	"fmt"

	sdkTypes "github.com/raito-io/sdk-go/types"
)

//go:generate go run github.com/vektra/mockery/v2 --name=UserClient --with-expecter --inpackage --replace-type github.com/raito-io/sdk-go/internal/schema=github.com/raito-io/sdk-go/types
type UserClient interface {
	GetUserByEmail(ctx context.Context, email string) (*sdkTypes.User, error)
	GetCurrentUser(ctx context.Context) (*sdkTypes.User, error)
}

type IdentityRepository struct {
	userClient UserClient

	// Cache
	usersByEmail map[string]*sdkTypes.User
}

func NewIdentityRepository(userClient UserClient) *IdentityRepository {
	return &IdentityRepository{
		userClient:   userClient,
		usersByEmail: make(map[string]*sdkTypes.User),
	}
}

func (r *IdentityRepository) GetUserByEmail(ctx context.Context, email string) (*sdkTypes.User, error) {
	if user, ok := r.usersByEmail[email]; ok {
		return user, nil
	}

	user, err := r.userClient.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user by email call: %w", err)
	}

	r.usersByEmail[email] = user

	return user, nil
}

func (r *IdentityRepository) GetCurrentUser(ctx context.Context) (*sdkTypes.User, error) {
	user, err := r.userClient.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("current user call: %w", err)
	}

	return user, nil
}
