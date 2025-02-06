package config

import (
	"os"

	"github.com/0987363/dns-failover/models"
	"gopkg.in/yaml.v2"
)

func LoadConfig(file string) (*models.Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.IpTest != nil {
		for _, domain := range config.Domains {
			if domain.IpTest == nil {
				domain.IpTest = config.IpTest
			}
		}
	}

	return &config, nil
}
