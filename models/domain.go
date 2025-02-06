package models

import "time"

type Domain struct {
	Name    string        `yaml:"name"`
	IPType  string        `yaml:"ip_type"`
	Proxied bool          `yaml:"proxied"`
	IPs     []string      `yaml:"ips"`
	IpTest  *IpTestConfig `yaml:"ip_test"`
}

type IpTestConfig struct {
	Interval time.Duration `yaml:"interval"`
	Sampling int           `yaml:"sampling"`
	Timeout  time.Duration `yaml:"timeout"`
}
type DomainRecord struct {
	Name    string
	IP      string
	IPType  string
	Proxied bool
}
