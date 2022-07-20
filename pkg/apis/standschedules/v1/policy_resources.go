/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

type AzureResourceType string

const (
	AzureResourceManagedMySQL   AzureResourceType = "mysql"
	AzureResourceVirtualMachine AzureResourceType = "vm"
)

type ResourcesSpec struct {
	// Azure contains an array of related azure resources.
	Azure []AzureResource `json:"azure"`
}

type AzureResource struct {
	// Type defines one of supported azure resource types.
	Type AzureResourceType `json:"type"`

	// ResourceGroupName defines resource group name for resource.
	ResourceGroupName string `json:"resourceGroupName"`

	// ResourceNameFilter defines regex filter for resource.
	ResourceNameFilter string `json:"resourceNameFilter"`

	// Priority specifies order in which resources will be started or shutdowned.
	Priority int64 `json:"priority"`
}
