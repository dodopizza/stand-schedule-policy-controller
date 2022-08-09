//go:build integration

package test

import (
	"context"
	"fmt"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
)

type (
	azureFixture struct {
		resources      []*azure.Resource
		startupErrors  map[string]error
		shutdownErrors map[string]error
	}
)

func (f *fixture) WithAzureResources(resources ...*azure.Resource) *fixture {
	f.azure.resources = resources
	return f
}

func (f *fixture) WithStartupFailed(resource *azure.Resource) *fixture {
	f.azure.startupErrors[f.azure.key(resource)] =
		fmt.Errorf("shutdown failure for %s", f.azure.key(resource))
	return f
}

func (f *fixture) WithShutdownFailed(resource *azure.Resource) *fixture {
	f.azure.shutdownErrors[f.azure.key(resource)] =
		fmt.Errorf("startup failure for %s", f.azure.key(resource))
	return f
}

func (az *azureFixture) List(_ context.Context, resourceType azure.ResourceType, resourceGroup string) ([]*azure.Resource, error) {
	var ret []*azure.Resource

	for _, resource := range az.resources {
		if resource.GetType() == resourceType && resource.GetResourceGroup() == resourceGroup {
			ret = append(ret, resource)
		}
	}

	return ret, nil
}

func (az *azureFixture) Shutdown(_ context.Context, resource *azure.Resource, _ bool) error {
	return az.shutdownErrors[az.key(resource)]
}

func (az *azureFixture) Startup(_ context.Context, resource *azure.Resource, _ bool) error {
	return az.startupErrors[az.key(resource)]
}

func (az *azureFixture) key(resource *azure.Resource) string {
	return fmt.Sprintf("%s/%s/%s", resource.GetType(), resource.GetResourceGroup(), resource.GetName())
}
