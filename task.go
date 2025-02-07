package main

import (
	"time"

	"github.com/0987363/dns-failover/models"
	"github.com/0987363/dns-failover/provider"

	log "github.com/sirupsen/logrus"
)

func ScheduledTask(provider provider.Provider, domainConfig *models.Domain) {
	ticker := time.NewTicker(domainConfig.IpTest.Interval)
	defer ticker.Stop()

	log.Debug("Immediate execution: ", time.Now().Format(time.RFC3339))
	executeTask(provider, domainConfig)

	for {
		select {
		case <-ticker.C:
			log.Debug("Start scheduled task:", time.Now().Format(time.RFC3339))
			executeTask(provider, domainConfig)
		}
	}
}

func executeTask(provider provider.Provider, domainConfig *models.Domain) {
	var bestPing *models.PingResult
	for _, ip := range domainConfig.IPs {
		res, err := Ping(ip, domainConfig.IpTest.Timeout, domainConfig.IpTest.Sampling)
		if err != nil {
			log.Errorf("Ping %s failed: %v", ip, err)
			continue
		}
		res.IP = ip
		log.Infof("Ping result: %+v", res)

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
	log.Infof("Start update dns: %+v", dr)
	if err := provider.UpdateDNS(dr); err != nil {
		log.Error("Update dns failed: ", err)
	}
}
