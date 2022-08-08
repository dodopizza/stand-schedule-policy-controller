//go:build integration

package test

import (
	"context"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
)

type (
	azureFixture struct{}
)

func (az *azureFixture) List(context.Context, azure.ResourceType, string) ([]*azure.Resource, error) {
	return nil, nil
}

func (az *azureFixture) Shutdown(context.Context, *azure.Resource, bool) error {
	return nil
}

func (az *azureFixture) Startup(context.Context, *azure.Resource, bool) error {
	return nil
}
