package raito

import (
	"context"
	"strings"

	"github.com/raito-io/sdk"
	"github.com/raito-io/sdk/services"
)

type DbtConfig struct {
	Domain    string
	ApiUser   string
	ApiSecret string

	URLOverride *string
}

func NewClient(ctx context.Context, config *DbtConfig) *sdk.RaitoClient {
	clientOptions := make([]func(options *sdk.ClientOptions), 0, 1)

	if config.URLOverride != nil {
		urlOverride := *config.URLOverride
		urlOverride = strings.TrimSuffix(urlOverride, "/")

		clientOptions = append(clientOptions, sdk.WithUrlOverride(urlOverride))
	}

	return sdk.NewClient(ctx, config.Domain, config.ApiUser, config.ApiSecret, clientOptions...)
}

func NewAccessProviderClient(client *sdk.RaitoClient) *services.AccessProviderClient {
	return client.AccessProvider()
}
