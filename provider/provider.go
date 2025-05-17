package provider

import (
	"errors"

	"github.com/0987363/dns-failover/models"
	"github.com/0987363/dns-failover/provider/cloudflare"
)

type Provider interface {
	UpdateDNS(string, []*models.DomainRecord) error
}

func Init(config *models.ProviderConfig) Provider {
	switch config.Name {
	case "cloudflare":
		return &cloudflare.CloudflareProvider{Key: config.Key, Email: config.Email}
	default:
		panic(errors.New("Unknown provider: " + config.Name))
	}
}
