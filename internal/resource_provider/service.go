package resource_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/raito-io/bexpression/utils"
	"github.com/raito-io/cli/base/resource_provider"
	"github.com/raito-io/golang-set/set"
	"github.com/raito-io/sdk-go/services"
	sdkTypes "github.com/raito-io/sdk-go/types"
	"github.com/raito-io/sdk-go/types/models"

	"cli-plugin-dbt/internal/array"
	"cli-plugin-dbt/internal/manifest"
	"cli-plugin-dbt/internal/resource_provider/types"
	"cli-plugin-dbt/internal/workerpool"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AccessProviderClient --with-expecter --inpackage --replace-type github.com/raito-io/sdk-go/internal/schema=github.com/raito-io/sdk-go/types
type AccessProviderClient interface {
	CreateAccessProvider(ctx context.Context, ap sdkTypes.AccessProviderInput) (*sdkTypes.AccessProvider, error)
	UpdateAccessProvider(ctx context.Context, id string, ap sdkTypes.AccessProviderInput, ops ...func(options *services.UpdateAccessProviderOptions)) (*sdkTypes.AccessProvider, error)
	DeleteAccessProvider(ctx context.Context, id string, ops ...func(options *services.UpdateAccessProviderOptions)) error
	ListAccessProviders(ctx context.Context, ops ...func(options *services.AccessProviderListOptions)) <-chan sdkTypes.ListItem[sdkTypes.AccessProvider]
}

const (
	dbtSource  = "dbt"
	lockReason = "locked by dbt"

	maxWorkerPoolSize = uint(4)
)

type DbtService struct {
	dataSourceId         string
	accessProviderClient AccessProviderClient
	manifestParser       manifest.Parser
	logger               hclog.Logger
}

func NewDbtService(config *resource_provider.UpdateResourceInput, accessProviderClient AccessProviderClient, manifestParser manifest.Parser, logger hclog.Logger) *DbtService {
	return &DbtService{
		dataSourceId:         config.DataSourceId,
		accessProviderClient: accessProviderClient,
		manifestParser:       manifestParser,
		logger:               logger,
	}
}

func (s *DbtService) RunDbt(ctx context.Context, dbtFile string) (uint32, uint32, uint32, uint32, error) {
	manifest, err := s.loadDbtFile(dbtFile)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("load file %s: %w", dbtFile, err)
	}

	source, grants, filters, masks, err := s.loadAccessProvidersFromManifest(manifest)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("load access providers from manifest: %w", err)
	}

	grantIds, filterIds, maskIds, apsToRemove, err := s.loadExistingAps(ctx, source, grants, filters, masks)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return s.createAndUpdateAccessProviders(ctx, grants, grantIds, masks, maskIds, filters, filterIds, apsToRemove)
}

func (s *DbtService) createAndUpdateAccessProviders(ctx context.Context, grants map[string]*sdkTypes.AccessProviderInput, grantIds map[string]string, masks map[string]*sdkTypes.AccessProviderInput, maskIds map[string]string, filters map[string]*sdkTypes.AccessProviderInput, filterIds map[string]string, apsToRemove set.Set[string]) (uint32, uint32, uint32, uint32, error) {
	numberOfChanges := len(grants) + len(masks) + len(filters) + len(apsToRemove)

	var addedResource, updatedResource, deletedResources, failures, totalChangedMade uint32

	logChannel := make(chan ResourceStatus) // channel will be true if ap is updated successfully.

	createOrUpdateAp := func(name string, apInput *sdkTypes.AccessProviderInput, apIds map[string]string) (err error) {
		create := false

		defer func() {
			if err != nil {
				logChannel <- ResourceStatusFailure
			} else if create {
				logChannel <- ResourceStatusCreated
			} else {
				logChannel <- ResourceStatusUpdated
			}
		}()

		if id, found := apIds[name]; found {
			s.logger.Debug(fmt.Sprintf("update access provider %q (%q)", name, id))

			_, updateErr := s.accessProviderClient.UpdateAccessProvider(ctx, id, *apInput, services.WithAccessProviderOverrideLocks())
			if updateErr != nil {
				return fmt.Errorf("update access provider %q (%q): %w", name, id, updateErr)
			}
		} else {
			s.logger.Debug(fmt.Sprintf("create access provider %q", name))
			create = true

			_, createErr := s.accessProviderClient.CreateAccessProvider(ctx, *apInput)
			if createErr != nil {
				return fmt.Errorf("create access provider %q: %w", name, createErr)
			}
		}

		return nil
	}

	var logWg = sync.WaitGroup{}
	logWg.Add(1)

	go func() {
		defer logWg.Done()

		for apUpdate := range logChannel {
			switch apUpdate {
			case ResourceStatusFailure:
				failures++
			case ResourceStatusCreated:
				addedResource++
			case ResourceStatusUpdated:
				updatedResource++
			case ResourceStatusDeleted:
				deletedResources++
			}

			totalChangedMade++

			s.logger.Info(fmt.Sprintf("updated %d of %d access providers. %d successful, %d failures", totalChangedMade, numberOfChanges, addedResource+updatedResource+deletedResources, failures))
		}
	}()

	workerPool := workerpool.NewWorkerPool(ctx, maxWorkerPoolSize)

	for key := range grants {
		grant := grants[key]
		name := key

		workerPool.Go(func() error {
			return createOrUpdateAp(name, grant, grantIds)
		})
	}

	for key := range masks {
		mask := masks[key]
		name := key

		workerPool.Go(func() error {
			return createOrUpdateAp(name, mask, maskIds)
		})
	}

	for key := range filters {
		filter := filters[key]
		name := key

		workerPool.Go(func() error {
			return createOrUpdateAp(name, filter, filterIds)
		})
	}

	for key := range apsToRemove {
		oldAp := key

		workerPool.Go(func() (err error) {
			defer func() {
				if err != nil {
					logChannel <- ResourceStatusFailure
				} else {
					logChannel <- ResourceStatusDeleted
				}
			}()

			s.logger.Debug(fmt.Sprintf("delete access provider %q", oldAp))

			err = s.accessProviderClient.DeleteAccessProvider(ctx, oldAp, services.WithAccessProviderOverrideLocks())
			if err != nil {
				return fmt.Errorf("delete access provider %q: %w", oldAp, err)
			}

			return nil
		})
	}

	err := workerPool.Wait()

	close(logChannel)
	logWg.Wait()

	if err != nil {
		return addedResource, updatedResource, deletedResources, failures, fmt.Errorf("worker pool errors: %w", err)
	}

	return addedResource, updatedResource, deletedResources, failures, nil
}

func (s *DbtService) loadExistingAps(ctx context.Context, source string, grants map[string]*sdkTypes.AccessProviderInput, filters map[string]*sdkTypes.AccessProviderInput, masks map[string]*sdkTypes.AccessProviderInput) (map[string]string, map[string]string, map[string]string, set.Set[string], error) {
	cancelCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	existingAps := s.accessProviderClient.ListAccessProviders(cancelCtx, services.WithAccessProviderListFilter(&sdkTypes.AccessProviderFilterInput{
		Source: utils.Ptr(source),
	}))

	grantIds := make(map[string]string)
	maskIds := make(map[string]string)
	filterIds := make(map[string]string)
	apsToRemove := set.NewSet[string]()

	for existingAp := range existingAps {
		if existingAp.HasError() {
			return nil, nil, nil, nil, fmt.Errorf("list access providers: %w", existingAp.GetError())
		}

		ap := existingAp.GetItem()
		switch ap.Action {
		case models.AccessProviderActionGrant:
			if _, found := grants[ap.Name]; found {
				if _, idFound := grantIds[ap.Name]; idFound {
					apsToRemove.Add(ap.Id) // Remove ap with same name
				} else {
					grantIds[ap.Name] = ap.Id
				}
			} else {
				apsToRemove.Add(ap.Id)
			}
		case models.AccessProviderActionFiltered:
			if _, found := filters[ap.Name]; found {
				if _, idFound := filterIds[ap.Name]; idFound {
					apsToRemove.Add(ap.Id) // Remove ap with same name
				} else {
					filterIds[ap.Name] = ap.Id
				}
			} else {
				apsToRemove.Add(ap.Id)
			}
		case models.AccessProviderActionMask:
			if _, found := masks[ap.Name]; found {
				if _, idFound := maskIds[ap.Name]; idFound {
					apsToRemove.Add(ap.Id) // Remove ap with same name
				} else {
					maskIds[ap.Name] = ap.Id
				}
			} else {
				apsToRemove.Add(ap.Id)
			}
		default:
			continue
		}
	}

	return grantIds, filterIds, maskIds, apsToRemove, nil
}

func (s *DbtService) loadDbtFile(dbtFilePath string) (*types.Manifest, error) {
	jsonBytes, err := os.ReadFile(dbtFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading dbt file: %w", err)
	}

	var result types.Manifest

	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing dbt file: %w", err)
	}

	return &result, nil
}

func (s *DbtService) loadAccessProvidersFromManifest(manifest *types.Manifest) (string, map[string]*sdkTypes.AccessProviderInput, map[string]*sdkTypes.AccessProviderInput, map[string]*sdkTypes.AccessProviderInput, error) {
	source := _source(manifest.Metadata.ProjectName)

	grants := make(map[string]*sdkTypes.AccessProviderInput)
	filters := make(map[string]*sdkTypes.AccessProviderInput)
	masks := make(map[string]*sdkTypes.AccessProviderInput)

	var err error

	defaultLocks := []sdkTypes.AccessProviderLockDataInput{
		{
			LockKey: sdkTypes.AccessProviderLockWhatlock,
			Details: &sdkTypes.AccessProviderLockDetailsInput{
				Reason: utils.Ptr(lockReason),
			},
		},
		{
			LockKey: sdkTypes.AccessProviderLockNamelock,
			Details: &sdkTypes.AccessProviderLockDetailsInput{
				Reason: utils.Ptr(lockReason),
			},
		},
	}

	supportedResourceTypes := set.NewSet("model", "seed", "snapshot")

	for i := range manifest.Nodes {
		if !supportedResourceTypes.Contains(manifest.Nodes[i].ResourceType) {
			continue
		}

		databaseName := manifest.Nodes[i].Database
		schemaName := manifest.Nodes[i].Schema
		modelName := manifest.Nodes[i].Name
		doName := fmt.Sprintf("%s.%s.%s", databaseName, schemaName, modelName)

		for grandIdx, grant := range manifest.Nodes[i].Meta.Raito.Grant {
			if _, found := grants[grant.Name]; !found {
				grants[grant.Name] = &sdkTypes.AccessProviderInput{
					Name:       &manifest.Nodes[i].Meta.Raito.Grant[grandIdx].Name,
					Action:     utils.Ptr(models.AccessProviderActionGrant),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					Source:     &source,
					Locks:      defaultLocks,
				}
			}

			grants[grant.Name].WhatDataObjects = append(grants[grant.Name].WhatDataObjects, sdkTypes.AccessProviderWhatInputDO{
				Permissions:       array.Map(grant.Permissions, func(i string) *string { return &i }),
				GlobalPermissions: array.Map(grant.GlobalPermissions, func(i string) *string { return &i }),
				DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
					{
						Fullname:   doName,
						Datasource: s.dataSourceId,
					},
				},
			})
		}

		for filterIdx, filter := range manifest.Nodes[i].Meta.Raito.Filter {
			if _, found := filters[filter.Name]; !found {
				filters[filter.Name] = &sdkTypes.AccessProviderInput{
					Name:       &manifest.Nodes[i].Meta.Raito.Filter[filterIdx].Name,
					Action:     utils.Ptr(models.AccessProviderActionFiltered),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					PolicyRule: &manifest.Nodes[i].Meta.Raito.Filter[filterIdx].PolicyRule,
					Source:     &source,
					WhatDataObjects: []sdkTypes.AccessProviderWhatInputDO{
						{
							DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
								{
									Fullname:   doName,
									Datasource: s.dataSourceId,
								},
							},
						},
					},
					Locks: defaultLocks,
				}
			} else {
				err = multierror.Append(err, fmt.Errorf("filter %s already exists", filter.Name))
			}
		}

		for columnIdx, column := range manifest.Nodes[i].Columns {
			if column.Meta.Raito.Mask == nil {
				continue
			}

			if mask, found := masks[column.Meta.Raito.Mask.Name]; found {
				if mask.Type != nil && column.Meta.Raito.Mask.Type != nil && *column.Meta.Raito.Mask.Type != *mask.Type {
					err = multierror.Append(err, fmt.Errorf("mask %s already exists with different type", column.Meta.Raito.Mask.Name))

					continue
				}

				isValid := true

				for _, dos := range mask.WhatDataObjects {
					for _, do := range dos.DataObjectByName {
						if !strings.HasPrefix(do.Fullname, doName) {
							err = multierror.Append(err, fmt.Errorf("mask %s can not be applied on multiple tables", column.Meta.Raito.Mask.Name))
							isValid = false

							break
						}
					}

					if !isValid {
						break
					}
				}

				if !isValid {
					continue
				}
			} else {
				masks[column.Meta.Raito.Mask.Name] = &sdkTypes.AccessProviderInput{
					Name:       &manifest.Nodes[i].Columns[columnIdx].Meta.Raito.Mask.Name,
					Action:     utils.Ptr(models.AccessProviderActionMask),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					Source:     &source,
					Type:       column.Meta.Raito.Mask.Type,
					Locks:      defaultLocks,
				}
			}

			masks[column.Meta.Raito.Mask.Name].WhatDataObjects = append(masks[column.Meta.Raito.Mask.Name].WhatDataObjects, sdkTypes.AccessProviderWhatInputDO{
				DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
					{
						Fullname:   fmt.Sprintf("%s.%s", doName, column.Name),
						Datasource: s.dataSourceId,
					},
				},
			})
		}
	}

	if err != nil {
		return source, nil, nil, nil, err
	}

	return source, grants, filters, masks, nil
}

func _source(projectName string) string {
	return fmt.Sprintf("%s-%s", dbtSource, projectName)
}
