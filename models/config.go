package models

type ProviderConfig struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type Config struct {
	ProviderConfig ProviderConfig `yaml:"provider"`
	IpTest         *IpTestConfig  `yaml:"ip_test"`
	Domains        []Domain       `yaml:"domains"`
}
