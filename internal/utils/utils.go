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

func GetFullnamePrefix(cfg *config.ConfigMap) string {
	prefix := cfg.GetString(constants.FullNamePrefixParameterName)

	if prefix == "" {
		return ""
	} else {
		return prefix + "."
	}
}
