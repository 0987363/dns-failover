package models

import "time"

type ProviderConfig struct {
	Name  string `yaml:"name"`
	Key   string `yaml:"key"`
	Email string `yaml:"email"`
}

type IpTestConfig struct {
	Interval time.Duration `yaml:"interval"`
	Sampling int           `yaml:"sampling"`
	Timeout  time.Duration `yaml:"timeout"`
}

type Domain struct {
	Name    string        `yaml:"name"`
	IPType  string        `yaml:"ip_type"`
	Proxied bool          `yaml:"proxied"`
	IPs     []string      `yaml:"ips"`
	IpTest  *IpTestConfig `yaml:"ip_test"`
}

type Config struct {
	ProviderConfig ProviderConfig `yaml:"provider"`
	IpTest         *IpTestConfig  `yaml:"ip_test"`
	Domains        []*Domain      `yaml:"domains"`
}
