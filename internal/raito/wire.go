//go:build wireinject
// +build wireinject

package raito

import "github.com/google/wire"

var Wired = wire.NewSet(
	NewClient,
	NewAccessProviderClient,
)
