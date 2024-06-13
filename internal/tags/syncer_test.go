package tags

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/bexpression/utils"
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/raito-io/cli-plugin-dbt/internal/constants"
	"github.com/raito-io/cli-plugin-dbt/internal/manifest"
)

func TestTagImportService_SyncTags(t *testing.T) {
	type args struct {
		config *tag.TagSyncConfig
	}
	tests := []struct {
		name        string
		args        args
		wantSources []string
		wantTags    []tag.TagImportObject
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "manifest1",
			args: args{
				config: &tag.TagSyncConfig{
					ConfigMap: &config.ConfigMap{
						Parameters: map[string]string{
							constants.ManifestParameterName: "testdata/manifest_1.json",
						},
					},
					DataSourceId: "datSourceId1",
				},
			},
			wantSources: []string{"dbt-dbt_bq_demo"},
			wantTags: []tag.TagImportObject{
				{
					DataObjectFullName: utils.Ptr("bq-demodata.dbt_company.new_customers"),
					Key:                "tag",
					StringValue:        "raito_tag_1",
					Source:             "dbt-dbt_bq_demo",
				},
				{
					DataObjectFullName: utils.Ptr("bq-demodata.dbt_company.new_customers"),
					Key:                "tag",
					StringValue:        "raito_tag_2",
					Source:             "dbt-dbt_bq_demo",
				},
				{
					DataObjectFullName: utils.Ptr("bq-demodata.dbt_company.new_customers.Email"),
					Key:                "tag",
					StringValue:        "raito_tag_2",
					Source:             "dbt-dbt_bq_demo",
				}, {
					DataObjectFullName: utils.Ptr("bq-demodata.dbt_company.new_customers.Email"),
					Key:                "tag",
					StringValue:        "raito_tag_3",
					Source:             "dbt-dbt_bq_demo",
				},
				{
					DataObjectFullName: utils.Ptr("bq-demodata.dbt_company.customers"),
					Key:                "tag",
					StringValue:        "raito_tag_5",
					Source:             "dbt-dbt_bq_demo",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifestParser := manifest.NewManifestParser()
			logger := hclog.NewNullLogger()
			tagsHandler := mocks.NewSimpleTagHandler(t, 1)

			tagSyncer := NewTagImportService(manifestParser, logger)

			gotSources, err := tagSyncer.SyncTags(context.Background(), tagsHandler, tt.args.config)

			if !tt.wantErr(t, err) {
				return
			}

			assert.ElementsMatchf(t, tt.wantSources, gotSources, "SyncTags() gotSources = %v, wantSources = %v", tt.wantSources, gotSources)
			assert.ElementsMatchf(t, tt.wantTags, tagsHandler.Tags, "SyncTags() gotTags = %v, wantTags = %v", tt.wantTags, tagsHandler.Tags)
		})
	}
}

func Test_loadTagsFromManifest(t *testing.T) {
	type args struct {
		manifestData *manifest.Manifest
	}
	tests := []struct {
		name        string
		args        args
		wantSources []string
		wantTags    []tag.TagImportObject
		wantErr     bool
	}{
		{
			name: "No tags",
			args: args{
				manifestData: &manifest.Manifest{
					Metadata: manifest.Metadata{
						ProjectName: "project-name-1",
					},
					Nodes: map[string]manifest.Node{
						"someNode": {
							Database:     "db",
							Schema:       "schema",
							Name:         "model1",
							ResourceType: "model",
							Tags:         nil,
						},
						"someNodeOtherNode": {
							Database:     "db",
							Schema:       "schema",
							Name:         "model2",
							ResourceType: "model",
							Tags:         nil,
							Columns: map[string]manifest.Column{
								"column1": {
									Name:        "column1",
									Description: "",
									Meta:        manifest.Meta{},
									DataType:    nil,
									Tags:        nil,
								},
							},
							Config: manifest.NodeConfig{
								Tags: nil,
							},
						},
					},
				},
			},
			wantSources: []string{"dbt-project-name-1"},
			wantTags:    []tag.TagImportObject{},
			wantErr:     false,
		},
		{
			name: "Found tags",
			args: args{
				manifestData: &manifest.Manifest{
					Metadata: manifest.Metadata{
						ProjectName: "project-name-2",
					},
					Nodes: map[string]manifest.Node{
						"someNode": {
							Database:     "db",
							Schema:       "schema",
							Name:         "model1",
							ResourceType: "model",
							Tags:         []string{"tag1", "tag2"},
						},
						"someNodeOtherNode": {
							Database:     "db",
							Schema:       "schema",
							Name:         "model2",
							ResourceType: "model",
							Tags:         []string{"tag2", "tag3"},
							Columns: map[string]manifest.Column{
								"column1": {
									Name:        "column1",
									Description: "",
									Meta:        manifest.Meta{},
									DataType:    nil,
									Tags:        []string{"tag3", "tag4"},
								},
							},
							Config: manifest.NodeConfig{
								Tags: []string{"tag5"},
							},
						},
						"NonSupportedNode": {
							Database:     "db",
							Schema:       "schema",
							Name:         "test",
							ResourceType: "test",
							Tags:         []string{"tag9"},
							Config: manifest.NodeConfig{
								Tags: []string{"tag10"},
							},
						},
					},
				},
			},
			wantSources: []string{"dbt-project-name-2"},
			wantTags: []tag.TagImportObject{
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model1"),
					Key:                "tag",
					StringValue:        "tag1",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model1"),
					Key:                "tag",
					StringValue:        "tag2",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model2"),
					Key:                "tag",
					StringValue:        "tag2",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model2"),
					Key:                "tag",
					StringValue:        "tag3",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model2"),
					Key:                "tag",
					StringValue:        "tag5",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model2.column1"),
					Key:                "tag",
					StringValue:        "tag3",
					Source:             "dbt-project-name-2",
				},
				{
					DataObjectFullName: utils.Ptr("prefix.db.schema.model2.column1"),
					Key:                "tag",
					StringValue:        "tag4",
					Source:             "dbt-project-name-2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagHandler := mocks.NewSimpleTagHandler(t, 1)

			got, err := loadTagsFromManifest(tt.args.manifestData, "prefix.", tagHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTagsFromManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.ElementsMatchf(t, got, tt.wantSources, "loadTagsFromManifest() return sources %v, want %v", got, tt.wantSources)

			assert.ElementsMatchf(t, tagHandler.Tags, tt.wantTags, "loadTagsFromManifest() created tags %v, want %v", tagHandler.Tags, tt)
		})
	}
}
