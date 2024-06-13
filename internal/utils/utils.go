package utils

import (
	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-dbt/internal/constants"
)

var logger hclog.Logger

func init() {
	logger = base.Logger()
}

func GetLogger() hclog.Logger {
	return logger
}

func GetFullnamePrefix(config *config.ConfigMap) string {
	prefix := config.GetString(constants.FullNamePrefixParameterName)

	if prefix == "" {
		return ""
	} else {
		return prefix + "."
	}
}
