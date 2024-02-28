//go:build wireinject
// +build wireinject

package resource_provider

import "github.com/google/wire"

var Wired = wire.NewSet(
	NewDbtService,
	ParseConfig,
	NewResourceSyncer,
)
