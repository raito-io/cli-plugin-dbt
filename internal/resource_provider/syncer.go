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
	service *DbtService
}

func NewResourceSyncer(service *DbtService) *ResourceSyncer {
	return &ResourceSyncer{
		service: service,
	}
}

func (r ResourceSyncer) UpdateResources(ctx context.Context, config *resource_provider.UpdateResourceInput) (*resource_provider.UpdateResourceResult, error) {
	addedResources, updatedResource, deletedResources, failures, err := r.service.RunDbt(ctx, config.ConfigMap.GetString(constants.ManifestParameterName), utils.GetFullnamePrefix(config.ConfigMap))
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
