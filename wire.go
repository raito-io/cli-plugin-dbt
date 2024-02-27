//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	resource_provider2 "github.com/raito-io/cli/base/resource_provider"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/sdk/services"

	"cli-plugin-dbt/internal/manifest"
	"cli-plugin-dbt/internal/raito"
	"cli-plugin-dbt/internal/resource_provider"
	"cli-plugin-dbt/internal/utils"
)

func InitializeResourceProviderSyncer(ctx context.Context, config *resource_provider2.UpdateResourceInput) (wrappers.ResourceProviderSyncer, func(), error) {
	wire.Build(
		resource_provider.Wired,
		raito.Wired,

		manifest.GlobalManifestParser,
		utils.GetLogger,

		wire.Bind(new(resource_provider.AccessProviderClient), new(*services.AccessProviderClient)),
		wire.Bind(new(wrappers.ResourceProviderSyncer), new(*resource_provider.ResourceSyncer)),
	)

	return nil, nil, nil
}
