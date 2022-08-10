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
	f.azure.startupErrors[resource.String()] =
		fmt.Errorf("shutdown failure for %s", resource)
	return f
}

func (f *fixture) WithShutdownFailed(resource *azure.Resource) *fixture {
	f.azure.shutdownErrors[resource.String()] =
		fmt.Errorf("startup failure for %s", resource)
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
	return az.shutdownErrors[resource.String()]
}

func (az *azureFixture) Startup(_ context.Context, resource *azure.Resource, _ bool) error {
	return az.startupErrors[resource.String()]
}
