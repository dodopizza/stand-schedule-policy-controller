package azure

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type (
	Client struct {
		Cred azcore.TokenCredential
		cfg  *Config
	}
	Interface interface {
	}
	Config struct {
		AuthType       string `env-required:"true" json:"auth_type" env:"AZURE_AUTH_TYPE"`
		SubscriptionId string `env-required:"true" json:"subscription_id" env:"AZURE_SUBSCRIPTION_ID"`
	}
)

func NewForConfig(cfg *Config) (*Client, error) {
	switch cfg.AuthType {
	case "default":
		return NewForDefaultAuth(cfg)
	case "msi":
		return NewForMsiAuth(cfg)
	default:
		return nil, errors.New("invalid access type specified")
	}
}

func NewForMsiAuth(cfg *Config) (*Client, error) {
	c, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, err
	}
	return NewClient(c, cfg)
}

func NewForDefaultAuth(cfg *Config) (*Client, error) {
	c, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return NewClient(c, cfg)
}

func NewClient(cred azcore.TokenCredential, cfg *Config) (*Client, error) {
	return &Client{cred, cfg}, nil
}
