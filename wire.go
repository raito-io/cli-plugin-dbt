//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	resource_provider2 "github.com/raito-io/cli/base/resource_provider"
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/sdk-go/services"

	"github.com/raito-io/cli-plugin-dbt/internal/manifest"
	"github.com/raito-io/cli-plugin-dbt/internal/raito"
	"github.com/raito-io/cli-plugin-dbt/internal/resource_provider"
	"github.com/raito-io/cli-plugin-dbt/internal/tags"
	"github.com/raito-io/cli-plugin-dbt/internal/utils"
)

func InitializeResourceProviderSyncer(ctx context.Context, config *resource_provider2.UpdateResourceInput) (wrappers.ResourceProviderSyncer, func(), error) {
	wire.Build(
		resource_provider.Wired,
		raito.Wired,

		manifest.GlobalManifestParser,
		utils.GetLogger,

		wire.Bind(new(resource_provider.AccessProviderClient), new(*services.AccessProviderClient)),
		wire.Bind(new(resource_provider.UserRepo), new(*raito.IdentityRepository)),
		wire.Bind(new(resource_provider.RoleClient), new(*services.RoleClient)),
		wire.Bind(new(wrappers.ResourceProviderSyncer), new(*resource_provider.ResourceSyncer)),
	)

	return nil, nil, nil
}

func InitializeTagSyncer(ctx context.Context, config *tag.TagSyncConfig) (wrappers.TagSyncer, func(), error) {
	wire.Build(
		tags.Wired,
		utils.GetLogger,
		manifest.GlobalManifestParser,

		wire.Bind(new(tags.Parser), new(manifest.Parser)),
		wire.Bind(new(wrappers.TagSyncer), new(*tags.TagImportService)),
	)

	return nil, nil, nil
}
