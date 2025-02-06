package provider

import (
	"errors"

	"github.com/0987363/dns-failover/models"
	"github.com/0987363/dns-failover/provider/cloudflare"
)

type Provider interface {
	UpdateDNS(*models.DomainRecord) error
}

func Init(config *models.ProviderConfig) Provider {
	switch config.Name {
	case "cloudflare":
		return &cloudflare.CloudflareProvider{Key: config.Key}
	default:
		panic(errors.New("Unknown provider: " + config.Name))
	}
}
