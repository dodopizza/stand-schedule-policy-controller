package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	ResourceType string

	Resource struct {
		id *arm.ResourceID
	}
)

const (
	ResourceManagedMySQL   = ResourceType("Microsoft.DBforMySQL/servers")
	ResourceVirtualMachine = ResourceType("Microsoft.Compute/virtualMachines")
)

func NewResource(rawId string) *Resource {
	id, _ := arm.ParseResourceID(rawId)

	return &Resource{
		id: id,
	}
}

func (r Resource) GetType() ResourceType {
	return ResourceType(r.id.ResourceType.Type)
}

func (r Resource) GetName() string {
	return r.id.Name
}

func (r Resource) GetResourceGroup() string {
	return r.id.ResourceGroupName
}

func From(api apis.AzureResourceType) (ResourceType, error) {
	switch api {
	case apis.AzureResourceManagedMySQL:
		return ResourceManagedMySQL, nil
	case apis.AzureResourceVirtualMachine:
		return ResourceVirtualMachine, nil
	default:
		return "", ErrUnsupportedType
	}
}
