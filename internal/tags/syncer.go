package tags

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-dbt/internal/constants"
	"github.com/raito-io/cli-plugin-dbt/internal/manifest"
	"github.com/raito-io/cli-plugin-dbt/internal/utils"
)

var _ wrappers.TagSyncer = (*TagImportService)(nil)

const TagKey = "tag"

//go:generate go run github.com/vektra/mockery/v2 --name=Parser --with-expecter
type Parser interface {
	LoadManifest(path string) (*manifest.Manifest, error)
}

type TagImportService struct {
	manifestParser Parser
	logger         hclog.Logger
}

func NewTagImportService(manifestParser Parser, logger hclog.Logger) *TagImportService {
	return &TagImportService{
		manifestParser: manifestParser,
		logger:         logger,
	}
}

func (t TagImportService) SyncTags(_ context.Context, tagsHandler wrappers.TagHandler, config *tag.TagSyncConfig) ([]string, error) {
	manifestFile := config.ConfigMap.GetString(constants.ManifestParameterName)

	manifestData, err := t.manifestParser.LoadManifest(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("load file %s: %w", manifestFile, err)
	}

	prefix := utils.GetFullnamePrefix(config.ConfigMap)

	return loadTagsFromManifest(manifestData, prefix, tagsHandler)
}

func loadTagsFromManifest(manifestData *manifest.Manifest, fullnamePrefix string, tagsHandler wrappers.TagHandler) ([]string, error) {
	supportedResourceTypes := set.NewSet("model", "seed", "snapshot")

	source := fmt.Sprintf("dbt-%s", manifestData.Metadata.ProjectName)

	for i := range manifestData.Nodes {
		if !supportedResourceTypes.Contains(manifestData.Nodes[i].ResourceType) {
			continue
		}

		databaseName := manifestData.Nodes[i].Database
		schemaName := manifestData.Nodes[i].Schema
		modelName := manifestData.Nodes[i].Name
		doName := fmt.Sprintf("%s%s.%s.%s", fullnamePrefix, databaseName, schemaName, modelName)

		doTags := set.NewSet[string](manifestData.Nodes[i].Tags...)
		doTags.Add(manifestData.Nodes[i].Config.Tags...)

		err := addTags(tagsHandler, doName, source, doTags)
		if err != nil {
			return nil, err
		}

		for columnName := range manifestData.Nodes[i].Columns {
			columnFullName := fmt.Sprintf("%s.%s", doName, columnName)
			columnTags := set.NewSet[string](manifestData.Nodes[i].Columns[columnName].Tags...)
			columnTags.Add(manifestData.Nodes[i].Columns[columnName].Config.Tags...)

			err = addTags(tagsHandler, columnFullName, source, columnTags)
			if err != nil {
				return nil, err
			}
		}
	}

	return []string{source}, nil
}

func addTags(tagsHandler wrappers.TagHandler, doFullName string, source string, tags set.Set[string]) error {
	for tagValue := range tags {
		err := tagsHandler.AddTags(&tag.TagImportObject{
			DataObjectFullName: &doFullName,
			Key:                TagKey,
			StringValue:        tagValue,
			Source:             source,
		})

		if err != nil {
			return fmt.Errorf("add tag %s to %s: %w", tagValue, doFullName, err)
		}
	}

	return nil
}
