package resource_provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/bexpression/utils"
	"github.com/raito-io/cli/base/resource_provider"
	"github.com/raito-io/golang-set/set"
	"github.com/raito-io/sdk-go/services"
	sdkTypes "github.com/raito-io/sdk-go/types"
	"github.com/raito-io/sdk-go/types/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-dbt/internal/manifest"
)

func TestDbtService_createAndUpdateAccessProviders(t *testing.T) {
	currentUser := sdkTypes.User{
		Id:          "CurrentUserId",
		Name:        "CurrentUser",
		Email:       ptr.String("currentUser@raito.io"),
		IsRaitoUser: true,
		Type:        "machine",
	}

	type fields struct {
		dataSourceId string
		setup        func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo)
	}
	type args struct {
		ctx         context.Context
		grants      map[string]*AccessProviderInput
		grantIds    map[string]string
		masks       map[string]*AccessProviderInput
		maskIds     map[string]string
		filters     map[string]*AccessProviderInput
		filterIds   map[string]string
		apsToRemove set.Set[string]
	}
	type result struct {
		added    uint32
		updated  uint32
		removed  uint32
		failures uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		result  result
		wantErr bool
	}{
		{
			name: "create and update grants",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("grantName"), Action: utils.Ptr(models.AccessProviderActionGrant)}).Return(&sdkTypes.AccessProvider{Name: "grantName", Id: "generatedGrantId1"}, nil).Once()
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "grantId2", sdkTypes.AccessProviderInput{Name: ptr.String("grantName2"), Action: utils.Ptr(models.AccessProviderActionGrant)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "grantName2"}, nil).Once()

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "generatedGrantId1", ownerRoleId, "owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "grantId2", ownerRoleId, currentUser.Id).Return(nil, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				grants: map[string]*AccessProviderInput{
					"grantName": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet("owner1"),
					},
					"grantName2": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName2"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet[string](),
					},
				},
				grantIds: map[string]string{"grantName2": "grantId2"},
			},
			result: result{
				added:    1,
				updated:  1,
				removed:  0,
				failures: 0,
			},
			wantErr: false,
		},
		{
			name: "create filters",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)}).Return(&sdkTypes.AccessProvider{Name: "filterName", Id: "filterId1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "filterId2", sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "filterName2"}, nil)

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "filterId1", ownerRoleId, "owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "filterId2", ownerRoleId, currentUser.Id).Return(nil, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				filters: map[string]*AccessProviderInput{
					"filterName1": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)}, Owners: set.NewSet("owner1")},
					"filterName2": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)}, Owners: set.NewSet[string]()},
				},
				filterIds: map[string]string{"filterName2": "filterId2"},
			},
			result: result{
				added:    1,
				updated:  1,
				removed:  0,
				failures: 0,
			},
			wantErr: false,
		}, {
			name: "create masks",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)}).Return(&sdkTypes.AccessProvider{Name: "maskName1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "maskId2", sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "maskName2"}, nil)

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "maskId2", ownerRoleId, "Owner1").Return(nil, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				masks: map[string]*AccessProviderInput{
					"maskName1": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)}, Owners: set.NewSet[string]()},
					"maskName2": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)}, Owners: set.NewSet[string]("Owner1")},
				},
				maskIds: map[string]string{"maskName2": "maskId2"},
			},
			result: result{
				added:    1,
				updated:  1,
				removed:  0,
				failures: 0,
			},
			wantErr: false,
		},
		{
			name: "remove access providers",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "maskId2", mock.Anything).Return(nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "filterId2", mock.Anything).Return(nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "grantId2", mock.Anything).Return(nil)
				},
			},
			args: args{
				ctx:         context.Background(),
				apsToRemove: set.NewSet("maskId2", "filterId2", "grantId2"),
			},
			result: result{
				added:    0,
				updated:  0,
				removed:  3,
				failures: 0,
			},
			wantErr: false,
		},
		{
			name: "successful update",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("grantName"), Action: utils.Ptr(models.AccessProviderActionGrant)}).Return(&sdkTypes.AccessProvider{Name: "grantName", Id: "grantId1"}, nil).Once()
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "grantId2", sdkTypes.AccessProviderInput{Name: ptr.String("grantName2"), Action: utils.Ptr(models.AccessProviderActionGrant)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "grantName2"}, nil).Once()
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)}).Return(&sdkTypes.AccessProvider{Name: "filterName", Id: "filterId1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "filterId2", sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "filterName2"}, nil)
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)}).Return(&sdkTypes.AccessProvider{Name: "maskName1", Id: "maskId1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "maskId2", sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "maskName2"}, nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "maskId3", mock.Anything).Return(nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "filterId3", mock.Anything).Return(nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "grantId3", mock.Anything).Return(nil)

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "maskId2", ownerRoleId, currentUser.Id).Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "grantId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "filterId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "maskId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "filterId2", ownerRoleId, currentUser.Id).Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "grantId2", ownerRoleId, currentUser.Id).Return(nil, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				grants: map[string]*AccessProviderInput{
					"grantName": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet("Owner1"),
					},
					"grantName2": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName2"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet[string](),
					},
				},
				grantIds: map[string]string{"grantName2": "grantId2"},
				filters: map[string]*AccessProviderInput{
					"filterName1": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)},
						Owners: set.NewSet("Owner1"),
					},
					"filterName2": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)},
						Owners: set.NewSet[string](),
					},
				},
				filterIds: map[string]string{"filterName2": "filterId2"},
				masks: map[string]*AccessProviderInput{
					"maskName1": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)},
						Owners: set.NewSet("Owner1"),
					},
					"maskName2": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)},
						Owners: set.NewSet[string](),
					},
				},
				maskIds:     map[string]string{"maskName2": "maskId2"},
				apsToRemove: set.NewSet("maskId3", "filterId3", "grantId3"),
			},
			result: result{
				added:    3,
				updated:  3,
				removed:  3,
				failures: 0,
			},
		},
		{
			name: "update with errors",
			fields: fields{
				dataSourceId: "dsId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("grantName"), Action: utils.Ptr(models.AccessProviderActionGrant)}).Return(&sdkTypes.AccessProvider{Name: "grantName", Id: "grantId1"}, nil).Once()
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "grantId2", sdkTypes.AccessProviderInput{Name: ptr.String("grantName2"), Action: utils.Ptr(models.AccessProviderActionGrant)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "grantName2"}, nil).Once()
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)}).Return(&sdkTypes.AccessProvider{Name: "filterName", Id: "filterId1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "filterId2", sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)}, mock.Anything).Return(nil, errors.New("error")).Once()
					apClientMock.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)}).Return(&sdkTypes.AccessProvider{Name: "maskName1", Id: "maskId1"}, nil)
					apClientMock.EXPECT().UpdateAccessProvider(mock.Anything, "maskId2", sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)}, mock.Anything).Return(&sdkTypes.AccessProvider{Name: "maskName2"}, nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "maskId3", mock.Anything).Return(nil)
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "filterId3", mock.Anything).Return(errors.New("some error")).Once()
					apClientMock.EXPECT().DeleteAccessProvider(mock.Anything, "grantId3", mock.Anything).Return(nil)

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "maskId2", ownerRoleId, currentUser.Id).Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "grantId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "filterId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "maskId1", ownerRoleId, "Owner1").Return(nil, nil)
					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "grantId2", ownerRoleId, currentUser.Id).Return(nil, nil)
				},
			},
			args: args{
				ctx: context.Background(),
				grants: map[string]*AccessProviderInput{
					"grantName": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet("Owner1"),
					},
					"grantName2": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("grantName2"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
						Owners: set.NewSet[string](),
					},
				},
				grantIds: map[string]string{"grantName2": "grantId2"},
				filters: map[string]*AccessProviderInput{
					"filterName1": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("filterName1"), Action: utils.Ptr(models.AccessProviderActionFiltered)},
						Owners: set.NewSet("Owner1"),
					},
					"filterName2": {
						Input:  sdkTypes.AccessProviderInput{Name: ptr.String("filterName2"), Action: utils.Ptr(models.AccessProviderActionFiltered)},
						Owners: set.NewSet[string](),
					},
				},
				filterIds: map[string]string{"filterName2": "filterId2"},
				masks: map[string]*AccessProviderInput{
					"maskName1": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("maskName1"), Action: utils.Ptr(models.AccessProviderActionMask)}, Owners: set.NewSet("Owner1")},
					"maskName2": {Input: sdkTypes.AccessProviderInput{Name: ptr.String("maskName2"), Action: utils.Ptr(models.AccessProviderActionMask)}, Owners: set.NewSet[string]()},
				},
				maskIds:     map[string]string{"maskName2": "maskId2"},
				apsToRemove: set.NewSet("maskId3", "filterId3", "grantId3"),
			},
			result: result{
				added:    3,
				updated:  2,
				removed:  2,
				failures: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, apMock, roleMock, userMock := createDbtService(t, tt.fields.dataSourceId)
			tt.fields.setup(apMock, roleMock, userMock)

			added, updated, removed, failures, err := s.createAndUpdateAccessProviders(tt.args.ctx, &currentUser, tt.args.grants, tt.args.grantIds, tt.args.masks, tt.args.maskIds, tt.args.filters, tt.args.filterIds, tt.args.apsToRemove)

			if (err != nil) != tt.wantErr {
				t.Errorf("createAndUpdateAccessProviders() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equalf(t, tt.result.added, added, "Expected %d added access providers, got %d", tt.result.added, added)
			assert.Equalf(t, tt.result.updated, updated, "Expected %d updated access providers, got %d", tt.result.updated, updated)
			assert.Equalf(t, tt.result.removed, removed, "Expected %d removed access providers, got %d", tt.result.removed, removed)
			assert.Equalf(t, tt.result.failures, failures, "Expected %d failures, got %d", tt.result.failures, failures)
		})
	}
}

func TestDbtService_loadExistingAps(t *testing.T) {
	type fields struct {
		dataSourceId string
		setup        func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo)
	}
	type args struct {
		ctx     context.Context
		grants  map[string]*AccessProviderInput
		filters map[string]*AccessProviderInput
		masks   map[string]*AccessProviderInput
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantGrantIds    map[string]string
		wantMaskIds     map[string]string
		wantFilterIds   map[string]string
		wantApsToRemove set.Set[string]
		wantErr         bool
	}{
		{
			name: "success",
			fields: fields{
				dataSourceId: "datasourceId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().ListAccessProviders(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f ...func(*services.AccessProviderListOptions)) <-chan sdkTypes.ListItem[sdkTypes.AccessProvider] {
						outputChannel := make(chan sdkTypes.ListItem[sdkTypes.AccessProvider], 1)
						go func() {
							defer close(outputChannel)

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider 1",
								Id:     "ap1",
								Action: models.AccessProviderActionGrant,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider 2",
								Id:     "ap2",
								Action: models.AccessProviderActionFiltered,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider Purpose",
								Id:     "purpose1",
								Action: models.AccessProviderActionPurpose,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider 3",
								Id:     "ap3",
								Action: models.AccessProviderActionMask,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider 4",
								Id:     "ap4",
								Action: models.AccessProviderActionGrant,
							})

						}()

						return outputChannel
					})
				},
			},
			args: args{
				ctx: context.Background(),
				grants: map[string]*AccessProviderInput{
					"access provider 1": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider 1"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
					},
				},
				filters: map[string]*AccessProviderInput{
					"access provider 2": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider 2"),
							Action: utils.Ptr(models.AccessProviderActionFiltered),
						},
					},
					"access provider 5": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider 5"),
							Action: utils.Ptr(models.AccessProviderActionFiltered),
						},
					},
				},
				masks: map[string]*AccessProviderInput{
					"access provider 3": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider 3"),
							Action: utils.Ptr(models.AccessProviderActionMask),
						},
					},
				},
			},
			wantGrantIds:    map[string]string{"access provider 1": "ap1"},
			wantFilterIds:   map[string]string{"access provider 2": "ap2"},
			wantMaskIds:     map[string]string{"access provider 3": "ap3"},
			wantApsToRemove: set.NewSet("ap4"),
			wantErr:         false,
		},
		{
			name: "multiple access providers with same name",
			fields: fields{
				dataSourceId: "datasourceId1",
				setup: func(apClientMock *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					apClientMock.EXPECT().ListAccessProviders(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f ...func(*services.AccessProviderListOptions)) <-chan sdkTypes.ListItem[sdkTypes.AccessProvider] {
						outputChannel := make(chan sdkTypes.ListItem[sdkTypes.AccessProvider], 1)
						go func() {
							defer close(outputChannel)

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap1",
								Action: models.AccessProviderActionGrant,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap2",
								Action: models.AccessProviderActionFiltered,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap3",
								Action: models.AccessProviderActionGrant,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap4",
								Action: models.AccessProviderActionFiltered,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap5",
								Action: models.AccessProviderActionMask,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "access provider with duplicated name",
								Id:     "ap6",
								Action: models.AccessProviderActionMask,
							})

						}()

						return outputChannel
					})
				},
			},
			args: args{
				ctx: context.Background(),
				grants: map[string]*AccessProviderInput{
					"access provider with duplicated name": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider with duplicated name"),
							Action: utils.Ptr(models.AccessProviderActionGrant),
						},
					},
				},
				filters: map[string]*AccessProviderInput{
					"access provider with duplicated name": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("access provider with duplicated name"),
							Action: utils.Ptr(models.AccessProviderActionFiltered),
						},
					},
					"new access provider": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("new access provider 5"),
							Action: utils.Ptr(models.AccessProviderActionFiltered),
						},
					},
				},
				masks: map[string]*AccessProviderInput{
					"access provider with duplicated name": {
						Input: sdkTypes.AccessProviderInput{
							Name:   ptr.String("aaccess provider with duplicated name"),
							Action: utils.Ptr(models.AccessProviderActionMask),
						},
					},
				},
			},
			wantGrantIds:    map[string]string{"access provider with duplicated name": "ap1"},
			wantFilterIds:   map[string]string{"access provider with duplicated name": "ap2"},
			wantMaskIds:     map[string]string{"access provider with duplicated name": "ap5"},
			wantApsToRemove: set.NewSet("ap3", "ap4", "ap6"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, apClientMock, roleMock, userMock := createDbtService(t, tt.fields.dataSourceId)
			tt.fields.setup(apClientMock, roleMock, userMock)

			got, got1, got2, got3, err := s.loadExistingAps(tt.args.ctx, "dpt-project", tt.args.grants, tt.args.filters, tt.args.masks)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadExistingAps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantGrantIds) {
				t.Errorf("loadExistingAps() got = %v, want %v", got, tt.wantGrantIds)
			}
			if !reflect.DeepEqual(got1, tt.wantFilterIds) {
				t.Errorf("loadExistingAps() got1 = %v, want %v", got1, tt.wantFilterIds)
			}
			if !reflect.DeepEqual(got2, tt.wantMaskIds) {
				t.Errorf("loadExistingAps() got2 = %v, want %v", got2, tt.wantMaskIds)
			}
			if !reflect.DeepEqual(got3, tt.wantApsToRemove) {
				t.Errorf("loadExistingAps() got3 = %v, want %v", got3, tt.wantApsToRemove)
			}
		})
	}
}

func TestDbtService_RunDbt(t *testing.T) {
	type fields struct {
		setup        func(client *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo)
		dataSourceId string
	}
	type args struct {
		ctx     context.Context
		dbtFile string
	}
	type result struct {
		added    uint32
		updated  uint32
		removed  uint32
		failures uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		result  result
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "manifest file 1",
			fields: fields{
				setup: func(client *MockAccessProviderClient, roleMock *MockRoleClient, userMock *MockUserRepo) {
					userMock.EXPECT().GetCurrentUser(mock.Anything).Return(&sdkTypes.User{
						Id:   "CurrentUserId",
						Name: "CurrentUser",
					}, nil).Once()

					userMock.EXPECT().GetUserByEmail(mock.Anything, "user1@raito.io").Return(&sdkTypes.User{
						Id:   "user1Id",
						Name: "User1",
					}, nil)

					userMock.EXPECT().GetUserByEmail(mock.Anything, "user2@raito.io").Return(&sdkTypes.User{
						Id:   "user2Id",
						Name: "User2",
					}, nil)

					roleMock.EXPECT().UpdateRoleAssigneesOnAccessProvider(mock.Anything, "apId1", ownerRoleId, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, apId string, roleId string, assignees ...string) (*sdkTypes.Role, error) {
						assert.ElementsMatch(t, []string{"user1Id", "user2Id"}, assignees)

						return nil, nil
					})

					client.EXPECT().ListAccessProviders(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f ...func(*services.AccessProviderListOptions)) <-chan sdkTypes.ListItem[sdkTypes.AccessProvider] {
						outputChannel := make(chan sdkTypes.ListItem[sdkTypes.AccessProvider])

						go func() {
							defer close(outputChannel)

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "sales_analysis_dbt",
								Id:     "apId1",
								Action: models.AccessProviderActionGrant,
							})

							outputChannel <- sdkTypes.NewListItemItem(&sdkTypes.AccessProvider{
								Name:   "another-ap",
								Id:     "apId2",
								Action: models.AccessProviderActionGrant,
							})
						}()

						return outputChannel
					})

					client.EXPECT().UpdateAccessProvider(mock.Anything, "apId1", sdkTypes.AccessProviderInput{
						Name:       utils.Ptr("sales_analysis_dbt"),
						Action:     utils.Ptr(models.AccessProviderActionGrant),
						WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
						Source:     utils.Ptr("dbt-dbt_bq_demo"),
						DataSource: utils.Ptr("dsId1"),
						WhatDataObjects: []sdkTypes.AccessProviderWhatInputDO{
							{
								GlobalPermissions: []*string{
									utils.Ptr("READ"),
								},
								Permissions: []*string{},
								DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
									{
										Fullname:   "prefix.bq-demodata.dbt_company.new_customers",
										Datasource: "dsId1",
									},
								},
							},
						},
						Locks: []sdkTypes.AccessProviderLockDataInput{
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
						},
					}, mock.Anything).Return(nil, nil).Once()

					client.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{
						Name:       utils.Ptr("country_filter_eu"),
						Action:     utils.Ptr(models.AccessProviderActionFiltered),
						WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
						Source:     utils.Ptr("dbt-dbt_bq_demo"),
						PolicyRule: utils.Ptr("Country IN (\"France\", \"Belgium\", \"Germany\")"),
						DataSource: utils.Ptr("dsId1"),
						WhatDataObjects: []sdkTypes.AccessProviderWhatInputDO{
							{
								DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
									{
										Fullname:   "prefix.bq-demodata.dbt_company.new_customers",
										Datasource: "dsId1",
									},
								},
							},
						},
						Locks: []sdkTypes.AccessProviderLockDataInput{
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
						},
					}).Return(&sdkTypes.AccessProvider{Id: "counter_filter_eu_id_1"}, nil).Once()

					client.EXPECT().CreateAccessProvider(mock.Anything, sdkTypes.AccessProviderInput{
						Name:       utils.Ptr("email_masking"),
						Action:     utils.Ptr(models.AccessProviderActionMask),
						WhatType:   utils.Ptr(sdkTypes.WhoAndWhatTypeStatic),
						Source:     utils.Ptr("dbt-dbt_bq_demo"),
						Type:       utils.Ptr("SHA256"),
						DataSource: utils.Ptr("dsId1"),
						WhatDataObjects: []sdkTypes.AccessProviderWhatInputDO{
							{
								DataObjectByName: []sdkTypes.AccessProviderWhatDoByNameInput{
									{
										Fullname:   "prefix.bq-demodata.dbt_company.new_customers.Email",
										Datasource: "dsId1",
									},
								},
							},
						},
						Locks: []sdkTypes.AccessProviderLockDataInput{
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
						},
					}).Return(&sdkTypes.AccessProvider{Id: "email_masking_id_2"}, nil).Once()

					client.EXPECT().DeleteAccessProvider(mock.Anything, "apId2", mock.Anything).Return(nil).Once()

				},
				dataSourceId: "dsId1",
			},
			args: args{
				ctx:     context.Background(),
				dbtFile: "testdata/manifest_1.json",
			},
			result: result{
				added:    2,
				updated:  1,
				removed:  1,
				failures: 0,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, accessProviderClientMock, roleMock, userMock := createDbtService(t, tt.fields.dataSourceId)

			tt.fields.setup(accessProviderClientMock, roleMock, userMock)

			added, updated, removed, failures, err := s.RunDbt(tt.args.ctx, tt.args.dbtFile, "prefix.")
			if !tt.wantErr(t, err, fmt.Sprintf("RunDbt(%v, %v)", tt.args.ctx, tt.args.dbtFile)) {
				return
			}
			assert.Equalf(t, tt.result.added, added, "Expected %d added access providers, got %d", tt.result.added, added)
			assert.Equalf(t, tt.result.updated, updated, "Expected %d updated access providers, got %d", tt.result.updated, updated)
			assert.Equalf(t, tt.result.removed, removed, "Expected %d removed access providers, got %d", tt.result.removed, removed)
			assert.Equalf(t, tt.result.failures, failures, "Expected %d failed access providers, got %d", tt.result.failures, failures)
		})
	}
}

func createDbtService(t *testing.T, dataSourceId string) (*DbtService, *MockAccessProviderClient, *MockRoleClient, *MockUserRepo) {
	t.Helper()

	apMock := NewMockAccessProviderClient(t)
	roleMock := NewMockRoleClient(t)
	userRepoMock := NewMockUserRepo(t)
	logger := hclog.NewNullLogger()
	manifestParser := manifest.NewManifestParser()

	service := NewDbtService(&resource_provider.UpdateResourceInput{DataSourceId: dataSourceId}, apMock, userRepoMock, roleMock, manifestParser, logger)

	return service, apMock, roleMock, userRepoMock
}
