package resource_provider

import (
	"github.com/raito-io/golang-set/set"
	sdkTypes "github.com/raito-io/sdk-go/types"
)

type ResourceStatus int

const (
	ResourceStatusFailure ResourceStatus = iota
	ResourceStatusCreated
	ResourceStatusUpdated
	ResourceStatusDeleted
)

type AccessProviderInput struct {
	Input  sdkTypes.AccessProviderInput
	Owners set.Set[string]
}
