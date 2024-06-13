//go:build wireinject
// +build wireinject

package raito

import (
	"github.com/google/wire"
	"github.com/raito-io/sdk-go/services"
)

var Wired = wire.NewSet(
	NewClient,
	NewAccessProviderClient,
	NewUserClient,
	NewRoleClient,
	NewIdentityRepository,

	wire.Bind(new(UserClient), new(*services.UserClient)),
)
