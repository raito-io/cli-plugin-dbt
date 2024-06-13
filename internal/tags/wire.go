//go:build wireinject
// +build wireinject

package tags

import "github.com/google/wire"

var Wired = wire.NewSet(
	NewTagImportService,
	NewTagSeparator,
)
