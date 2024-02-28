package resource_provider

import (
	"github.com/raito-io/cli/base/resource_provider"

	"cli-plugin-dbt/internal/raito"
)

func ParseConfig(input *resource_provider.UpdateResourceInput) *raito.DbtConfig {
	return &raito.DbtConfig{
		Domain:      input.Domain,
		ApiUser:     input.Credentials.Username,
		ApiSecret:   input.Credentials.Password,
		URLOverride: input.UrlOverride,
	}
}
