package azure

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

type (
	Config struct {
		AuthType       string `env-required:"true" json:"auth_type" env:"AZURE_AUTH_TYPE"`
		SubscriptionId string `env-required:"true" json:"subscription_id" env:"AZURE_SUBSCRIPTION_ID"`
	}
	client struct {
		cred  azcore.TokenCredential
		cfg   *Config
		mysql *armmysql.ServersClient
		vms   *armcompute.VirtualMachinesClient
	}
)

func NewForConfig(cfg *Config) (Interface, error) {
	switch cfg.AuthType {
	case "default":
		return NewForDefaultAuth(cfg)
	case "msi":
		return NewForMsiAuth(cfg)
	default:
		return nil, errors.New("invalid access type specified")
	}
}

func NewForMsiAuth(cfg *Config) (Interface, error) {
	c, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, err
	}
	return NewClient(c, cfg)
}

func NewForDefaultAuth(cfg *Config) (Interface, error) {
	c, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return NewClient(c, cfg)
}

func NewClient(cred azcore.TokenCredential, cfg *Config) (Interface, error) {
	client := &client{
		cred: cred,
		cfg:  cfg,
	}

	mysql, err := armmysql.NewServersClient(client.cfg.SubscriptionId, client.cred, nil)
	if err != nil {
		return nil, err
	}
	client.mysql = mysql

	vms, err := armcompute.NewVirtualMachinesClient(client.cfg.SubscriptionId, client.cred, nil)
	if err != nil {
		return nil, err
	}
	client.vms = vms

	return client, nil
}
