package provider

type DataProvider interface {
	SetLatestVersion() error
	GetServiceSchemaMap() (map[string]ServiceSchema, error)
	GetResultFileName() string
}

type ServiceSchema struct {
	APIVersion         string
	ServiceId          string
	ServiceFullName    string
	EndpointPrefix     string
	RegionTableVersion string
	Operations         []string
}
