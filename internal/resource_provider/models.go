package resource_provider

type DbtConfig struct {
	Domain    string
	ApiUser   string
	ApiSecret string

	DataSourceId string

	URLOverride *string
}

type ResourceStatus int

const (
	ResourceStatusFailure ResourceStatus = iota
	ResourceStatusCreated
	ResourceStatusUpdated
	ResourceStatusDeleted
)
