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

	"github.com/raito-io/cli-plugin-dbt/internal/array"
	"github.com/raito-io/cli-plugin-dbt/internal/manifest"
	"github.com/raito-io/cli-plugin-dbt/internal/resource_provider/types"
	"github.com/raito-io/cli-plugin-dbt/internal/workerpool"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AccessProviderClient --with-expecter --inpackage --replace-type github.com/raito-io/sdk-go/internal/schema=github.com/raito-io/sdk-go/types
type AccessProviderClient interface {
	CreateAccessProvider(ctx context.Context, ap sdkTypes.AccessProviderInput) (*sdkTypes.AccessProvider, error)
	UpdateAccessProvider(ctx context.Context, id string, ap sdkTypes.AccessProviderInput, ops ...func(options *services.UpdateAccessProviderOptions)) (*sdkTypes.AccessProvider, error)
	DeleteAccessProvider(ctx context.Context, id string, ops ...func(options *services.UpdateAccessProviderOptions)) error
	ListAccessProviders(ctx context.Context, ops ...func(options *services.AccessProviderListOptions)) <-chan sdkTypes.ListItem[sdkTypes.AccessProvider]
}

//go:generate go run github.com/vektra/mockery/v2 --name=RoleClient --with-expecter --inpackage --replace-type github.com/raito-io/sdk-go/internal/schema=github.com/raito-io/sdk-go/types
type RoleClient interface {
	UpdateRoleAssigneesOnAccessProvider(ctx context.Context, accessProviderId string, roleId string, assignees ...string) (*sdkTypes.Role, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=UserRepo --with-expecter --inpackage --replace-type github.com/raito-io/sdk-go/internal/schema=github.com/raito-io/sdk-go/types
type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*sdkTypes.User, error)
}

const (
	dbtSource  = "dbt"
	lockReason = "locked by dbt"

	ownerRoleId = "OwnerRole"

	maxWorkerPoolSize = uint(4)
)

type DbtService struct {
	dataSourceId         string
	accessProviderClient AccessProviderClient
	userRepo             UserRepo
	roleClient           RoleClient
	manifestParser       manifest.Parser
	logger               hclog.Logger
}

func NewDbtService(config *resource_provider.UpdateResourceInput, accessProviderClient AccessProviderClient, userRepo UserRepo, roleClient RoleClient, manifestParser manifest.Parser, logger hclog.Logger) *DbtService {
	return &DbtService{
		dataSourceId:         config.DataSourceId,
		accessProviderClient: accessProviderClient,
		userRepo:             userRepo,
		roleClient:           roleClient,
		manifestParser:       manifestParser,
		logger:               logger,
	}
}

func (s *DbtService) RunDbt(ctx context.Context, dbtFile string, fullnamePrefix string) (uint32, uint32, uint32, uint32, error) {
	manifestData, err := s.loadDbtFile(dbtFile)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("load file %s: %w", dbtFile, err)
	}

	source, grants, filters, masks, err := s.loadAccessProvidersFromManifest(ctx, manifestData, fullnamePrefix)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("load access providers from manifest: %w", err)
	}

	grantIds, filterIds, maskIds, apsToRemove, err := s.loadExistingAps(ctx, source, grants, filters, masks)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return s.createAndUpdateAccessProviders(ctx, grants, grantIds, masks, maskIds, filters, filterIds, apsToRemove)
}

func (s *DbtService) createAndUpdateAccessProviders(ctx context.Context, grants map[string]*AccessProviderInput, grantIds map[string]string, masks map[string]*AccessProviderInput, maskIds map[string]string, filters map[string]*AccessProviderInput, filterIds map[string]string, apsToRemove set.Set[string]) (uint32, uint32, uint32, uint32, error) {
	numberOfChanges := len(grants) + len(masks) + len(filters) + len(apsToRemove)

	var addedResource, updatedResource, deletedResources, failures, totalChangedMade uint32

	logChannel := make(chan ResourceStatus) // channel will be true if ap is updated successfully.

	createOrUpdateAp := func(name string, apInput *AccessProviderInput, apIds map[string]string) (err error) {
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

		var id string
		var found bool

		if id, found = apIds[name]; found {
			s.logger.Debug(fmt.Sprintf("update access provider %q (%q)", name, id))

			_, updateErr := s.accessProviderClient.UpdateAccessProvider(ctx, id, apInput.Input, services.WithAccessProviderOverrideLocks())
			if updateErr != nil {
				return fmt.Errorf("update access provider %q (%q): %w", name, id, updateErr)
			}
		} else {
			s.logger.Debug(fmt.Sprintf("create access provider %q", name))
			create = true

			ap, createErr := s.accessProviderClient.CreateAccessProvider(ctx, apInput.Input)
			if createErr != nil {
				return fmt.Errorf("create access provider %q: %w", name, createErr)
			}

			id = ap.Id
		}

		if len(apInput.Owners) > 0 {
			s.logger.Debug(fmt.Sprintf("update owners for access provider %q (%q)", name, id))

			_, ownerUpdateErr := s.roleClient.UpdateRoleAssigneesOnAccessProvider(ctx, id, ownerRoleId, apInput.Owners.Slice()...)
			if ownerUpdateErr != nil {
				return fmt.Errorf("update owners for access provider %q (%q): %w", name, id, ownerUpdateErr)
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

func (s *DbtService) loadExistingAps(ctx context.Context, source string, grants map[string]*AccessProviderInput, filters map[string]*AccessProviderInput, masks map[string]*AccessProviderInput) (map[string]string, map[string]string, map[string]string, set.Set[string], error) {
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

func (s *DbtService) loadAccessProvidersFromManifest(ctx context.Context, manifestData *types.Manifest, fullnamePrefix string) (string, map[string]*AccessProviderInput, map[string]*AccessProviderInput, map[string]*AccessProviderInput, error) {
	source := _source(manifestData.Metadata.ProjectName)

	grants := make(map[string]*AccessProviderInput)
	filters := make(map[string]*AccessProviderInput)
	masks := make(map[string]*AccessProviderInput)

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

	for i := range manifestData.Nodes {
		if !supportedResourceTypes.Contains(manifestData.Nodes[i].ResourceType) {
			continue
		}

		databaseName := manifestData.Nodes[i].Database
		schemaName := manifestData.Nodes[i].Schema
		modelName := manifestData.Nodes[i].Name
		doName := fmt.Sprintf("%s%s.%s.%s", fullnamePrefix, databaseName, schemaName, modelName)

		s.parseGrants(ctx, manifestData, i, grants, source, defaultLocks, doName)

		fErr := s.parseFilters(ctx, manifestData, i, filters, source, doName, defaultLocks)
		if fErr != nil {
			err = multierror.Append(err, fmt.Errorf("parse filters: %w", fErr))
		}

		mErr := s.parseMasks(ctx, manifestData, i, masks, doName, source, defaultLocks)
		if mErr != nil {
			err = multierror.Append(err, fmt.Errorf("parse masks: %w", mErr))
		}
	}

	if err != nil {
		return source, nil, nil, nil, err
	}

	return source, grants, filters, masks, nil
}

func (s *DbtService) parseMasks(ctx context.Context, manifestData *types.Manifest, i string, masks map[string]*AccessProviderInput, doName string, source string, defaultLocks []sdkTypes.AccessProviderLockDataInput) error {
	var err error

	for columnIdx, column := range manifestData.Nodes[i].Columns {
		if column.Meta.Raito.Mask == nil {
			continue
		}

		if mask, found := masks[column.Meta.Raito.Mask.Name]; found {
			if mask.Input.Type != nil && column.Meta.Raito.Mask.Type != nil && *column.Meta.Raito.Mask.Type != *mask.Input.Type {
				err = multierror.Append(err, fmt.Errorf("mask %s already exists with different type", column.Meta.Raito.Mask.Name))

				continue
			}

			isValid := true

			for _, dos := range mask.Input.WhatDataObjects {
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
			masks[column.Meta.Raito.Mask.Name] = &AccessProviderInput{
				Input: sdkTypes.AccessProviderInput{
					Name:       &manifestData.Nodes[i].Columns[columnIdx].Meta.Raito.Mask.Name,
					Action:     utils.Ptr(models.AccessProviderActionMask),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					Source:     &source,
					Type:       column.Meta.Raito.Mask.Type,
					Locks:      defaultLocks,
				},
				Owners: set.NewSet[string](),
			}
		}

		masks[column.Meta.Raito.Mask.Name].Input.WhatDataObjects = append(masks[column.Meta.Raito.Mask.Name].Input.WhatDataObjects, sdkTypes.AccessProviderWhatInputDO{
			DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
				{
					Fullname:   fmt.Sprintf("%s.%s", doName, column.Name),
					Datasource: s.dataSourceId,
				},
			},
		})

		ownerErr := s.handleOwners(ctx, masks[column.Meta.Raito.Mask.Name], column.Meta.Raito.Mask.Owners)
		if ownerErr != nil {
			s.logger.Warn(fmt.Sprintf("handle owners for mask %s: %v", column.Meta.Raito.Mask.Name, ownerErr))
		}
	}

	return err
}

func (s *DbtService) parseFilters(ctx context.Context, manifestData *types.Manifest, i string, filters map[string]*AccessProviderInput, source string, doName string, defaultLocks []sdkTypes.AccessProviderLockDataInput) error {
	var err error

	for filterIdx, filter := range manifestData.Nodes[i].Meta.Raito.Filter {
		if _, found := filters[filter.Name]; !found {
			filters[filter.Name] = &AccessProviderInput{
				Input: sdkTypes.AccessProviderInput{
					Name:       &manifestData.Nodes[i].Meta.Raito.Filter[filterIdx].Name,
					Action:     utils.Ptr(models.AccessProviderActionFiltered),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					PolicyRule: &manifestData.Nodes[i].Meta.Raito.Filter[filterIdx].PolicyRule,
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
				},
				Owners: set.NewSet[string](),
			}

			ownerErr := s.handleOwners(ctx, filters[filter.Name], filter.Owners)
			if ownerErr != nil {
				s.logger.Warn(fmt.Sprintf("handle owners for filter %s: %v", filter.Name, ownerErr))
			}
		} else {
			err = multierror.Append(err, fmt.Errorf("filter %s already exists", filter.Name))
		}
	}

	return err
}

func (s *DbtService) parseGrants(ctx context.Context, manifestData *types.Manifest, i string, grants map[string]*AccessProviderInput, source string, defaultLocks []sdkTypes.AccessProviderLockDataInput, doName string) {
	for grandIdx, grant := range manifestData.Nodes[i].Meta.Raito.Grant {
		if _, found := grants[grant.Name]; !found {
			grants[grant.Name] = &AccessProviderInput{
				Owners: set.NewSet[string](),
				Input: sdkTypes.AccessProviderInput{
					Name:       &manifestData.Nodes[i].Meta.Raito.Grant[grandIdx].Name,
					Action:     utils.Ptr(models.AccessProviderActionGrant),
					WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
					DataSource: &s.dataSourceId,
					Source:     &source,
					Locks:      defaultLocks,
				},
			}
		}

		grants[grant.Name].Input.WhatDataObjects = append(grants[grant.Name].Input.WhatDataObjects, sdkTypes.AccessProviderWhatInputDO{
			Permissions:       array.Map(grant.Permissions, func(i string) *string { return &i }),
			GlobalPermissions: array.Map(grant.GlobalPermissions, func(i string) *string { return &i }),
			DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
				{
					Fullname:   doName,
					Datasource: s.dataSourceId,
				},
			},
		})

		err := s.handleOwners(ctx, grants[grant.Name], grant.Owners)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("handle owners for grant %s: %v", grant.Name, err))
		}
	}
}

func (s *DbtService) handleOwners(ctx context.Context, ap *AccessProviderInput, owners []string) error {
	if len(owners) > 0 {
		users, ownerErr := s.getIdsOfUsers(ctx, owners...)
		if ownerErr != nil {
			return fmt.Errorf("get ids of users %v: %w", owners, ownerErr)
		}

		ap.Owners.Add(users...)

		ap.Input.Locks = append(ap.Input.Locks, sdkTypes.AccessProviderLockDataInput{
			LockKey: sdkTypes.AccessProviderLockOwnerlock,
			Details: &sdkTypes.AccessProviderLockDetailsInput{
				Reason: utils.Ptr(lockReason),
			},
		})
	}

	return nil
}

func (s *DbtService) getIdsOfUsers(ctx context.Context, emailAddresses ...string) ([]string, error) {
	result := make([]string, 0, len(emailAddresses))
	var err error

	for i := range emailAddresses {
		u, uErr := s.userRepo.GetUserByEmail(ctx, emailAddresses[i])
		if uErr != nil {
			err = multierror.Append(err, fmt.Errorf("get user by email %s: %w", emailAddresses[i], uErr))
			continue
		}

		result = append(result, u.Id)
	}

	return result, err
}

func _source(projectName string) string {
	return fmt.Sprintf("%s-%s", dbtSource, projectName)
}
