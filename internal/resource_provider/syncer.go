package resource_provider

import (
	"context"
	"fmt"

	"github.com/raito-io/cli/base/resource_provider"
	"github.com/raito-io/cli/base/wrappers"

	"cli-plugin-dbt/internal/constants"
	"cli-plugin-dbt/internal/utils"
)

var _ wrappers.ResourceProviderSyncer = (*ResourceSyncer)(nil)

type ResourceSyncer struct {
}

func (r ResourceSyncer) UpdateResources(ctx context.Context, config *resource_provider.UpdateResourceInput) (*resource_provider.UpdateResourceResult, error) {
	dbtConfig := DbtConfig{
		Domain:       config.Domain,
		ApiUser:      config.Credentials.Username,
		ApiSecret:    config.Credentials.Password,
		DataSourceId: config.DataSourceId,
		URLOverride:  config.UrlOverride,
	}

	service := NewDbtService(ctx, &dbtConfig, utils.GetLogger())

	addedResources, updatedResource, deletedResources, failures, err := service.RunDbt(ctx, config.ConfigMap.GetString(constants.ManifestParameterName))
	if err != nil {
		return nil, fmt.Errorf("running dbt: %w", err)
	}

	return &resource_provider.UpdateResourceResult{
		AddedObjects:   int32(addedResources),
		UpdatedObjects: int32(updatedResource),
		DeletedObjects: int32(deletedResources),
		Failures:       int32(failures),
	}, nil
}
