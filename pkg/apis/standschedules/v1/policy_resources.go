/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

type AzureResourceType string
type AzureResourceList []AzureResource

const (
	AzureResourceManagedMySQL   AzureResourceType = "mysql"
	AzureResourceVirtualMachine AzureResourceType = "vm"
)

type ResourcesSpec struct {
	// Azure contains an array of related azure resources.
	Azure AzureResourceList `json:"azure,omitempty"`
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

func (l AzureResourceList) Len() int {
	return len(l)
}

func (l AzureResourceList) Less(i, j int) bool {
	return l[i].Priority < l[j].Priority
}

func (l AzureResourceList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
