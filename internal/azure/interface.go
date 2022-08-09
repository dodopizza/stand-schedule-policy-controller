package azure

import (
	"context"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

type (
	Interface interface {
		List(ctx context.Context, resourceType ResourceType, resourceGroup string) ([]*Resource, error)
		Shutdown(ctx context.Context, resource *Resource, wait bool) error
		Startup(ctx context.Context, resource *Resource, wait bool) error
	}
)

var (
	ErrUnsupportedType = errors.New("unsupported type specified")
)

const (
	PollerInterval = time.Second * 10
)

func (c *client) List(ctx context.Context, resourceType ResourceType, resourceGroup string) ([]*Resource, error) {
	switch resourceType {
	case ResourceManagedMySQL:
		pager := c.mysql.NewListByResourceGroupPager(resourceGroup, nil)
		return List(ctx, pager,
			func(pager armmysql.ServersClientListByResourceGroupResponse) []*armmysql.Server {
				return pager.Value
			},
			func(s *armmysql.Server) *Resource { return NewResource(*s.ID) },
		)
	case ResourceVirtualMachine:
		pager := c.vms.NewListPager(resourceGroup, nil)
		return List(ctx, pager,
			func(pager armcompute.VirtualMachinesClientListResponse) []*armcompute.VirtualMachine {
				return pager.Value
			},
			func(vm *armcompute.VirtualMachine) *Resource { return NewResource(*vm.ID) },
		)
	default:
		return nil, ErrUnsupportedType
	}
}

func (c *client) Shutdown(ctx context.Context, resource *Resource, wait bool) error {
	switch resource.GetType() {
	case ResourceManagedMySQL:
		return ExecuteOperation(ctx, resource, wait,
			func(ctx context.Context, resource *Resource) (*runtime.Poller[armmysql.ServersClientStopResponse], error) {
				return c.mysql.BeginStop(ctx, resource.GetResourceGroup(), resource.GetName(), nil)
			})
	case ResourceVirtualMachine:
		return ExecuteOperation(ctx, resource, wait,
			func(ctx context.Context, resource *Resource) (*runtime.Poller[armcompute.VirtualMachinesClientDeallocateResponse], error) {
				return c.vms.BeginDeallocate(ctx, resource.GetResourceGroup(), resource.GetName(), nil)
			})
	default:
		return ErrUnsupportedType
	}
}

func (c *client) Startup(ctx context.Context, resource *Resource, wait bool) error {
	switch resource.GetType() {
	case ResourceManagedMySQL:
		return ExecuteOperation(ctx, resource, wait,
			func(ctx context.Context, resource *Resource) (*runtime.Poller[armmysql.ServersClientStartResponse], error) {
				return c.mysql.BeginStart(ctx, resource.GetResourceGroup(), resource.GetName(), nil)
			})
	case ResourceVirtualMachine:
		return ExecuteOperation(ctx, resource, wait,
			func(ctx context.Context, resource *Resource) (*runtime.Poller[armcompute.VirtualMachinesClientStartResponse], error) {
				return c.vms.BeginStart(ctx, resource.GetResourceGroup(), resource.GetName(), nil)
			})
	default:
		return ErrUnsupportedType
	}
}

func List[T any, TResource any](
	ctx context.Context,
	pager *runtime.Pager[T],
	values func(pager T) []*TResource,
	create func(*TResource) *Resource,
) (ret []*Resource, err error) {
	for {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, resource := range values(page) {
			ret = append(ret, create(resource))
		}
		if !pager.More() {
			break
		}
	}
	return
}

func ExecuteOperation[T any](
	ctx context.Context,
	resource *Resource,
	wait bool,
	poller func(context.Context, *Resource) (*runtime.Poller[T], error),
) error {
	p, err := poller(ctx, resource)
	if err != nil {
		return err
	}

	if !wait {
		return nil
	}

	_, err = p.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: PollerInterval})
	return err
}
