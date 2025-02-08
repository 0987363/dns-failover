package main

import (
	"time"

	"github.com/0987363/dns-failover/models"
	"github.com/0987363/dns-failover/provider"

	log "github.com/sirupsen/logrus"
)

func ScheduledTask(provider provider.Provider, domainConfig *models.Domain) {
	log.Infof("Immediate execution: %+v, test:%+v\n", domainConfig, domainConfig.IpTest)
	executeTask(provider, domainConfig)

	ticker := time.NewTicker(domainConfig.IpTest.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Infof("Start scheduled task: %+v, test:%+v\n", domainConfig, domainConfig.IpTest)
			executeTask(provider, domainConfig)
		}
	}
}

func executeTask(provider provider.Provider, domainConfig *models.Domain) {
	var bestPing *models.PingResult
	for _, ip := range domainConfig.IPs {
		res, err := Ping(ip, domainConfig.IpTest.Timeout, domainConfig.IpTest.Sampling)
		if err != nil {
			log.Errorf("Ping %s failed: %v\n", ip, err)
			continue
		}
		res.IP = ip
		log.Infof("Ping result: %+v\n", res)

		if bestPing == nil || res.Quality < bestPing.Quality {
			bestPing = res
		}
	}

	if bestPing == nil {
		log.Infof("No available IPs for domain %s\n", domainConfig.Name)
		return
	}

	dr := &models.DomainRecord{
		Name:    domainConfig.Name,
		IP:      bestPing.IP,
		IPType:  domainConfig.IPType,
		Proxied: domainConfig.Proxied,
	}
	log.Infof("Start update dns: %+v\n", dr)
	if err := provider.UpdateDNS(dr); err != nil {
		log.Error("Update dns failed: ", err)
	}
}
