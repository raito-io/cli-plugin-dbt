package resource_provider

type ResourceStatus int

const (
	ResourceStatusFailure ResourceStatus = iota
	ResourceStatusCreated
	ResourceStatusUpdated
	ResourceStatusDeleted
)
