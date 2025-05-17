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
	bestIP := []string{}
	for _, ips := range domainConfig.IPs {
		cleanIPs := models.SplitCommaSeparatedIPs(ips)
		if len(cleanIPs) == 0 {
			continue
		}

		var bestPing *models.PingResult
		for _, ip := range cleanIPs {
			res, err := Ping(ip, domainConfig.IpTest.Timeout, domainConfig.IpTest.Sampling)
			if err == nil {
				res.IP = ip
				log.Infof("Found best ip: %+v\n", res)

				if bestPing == nil || res.Quality < bestPing.Quality {
					bestPing = res
				}
			}
		}
		if bestPing != nil {
			bestIP = append(bestIP, bestPing.IP)
		}
	}
	log.Infof("Found best ip: %+v\n", bestIP)

	drs := []*models.DomainRecord{}
	for _, ip := range bestIP {
		dr := &models.DomainRecord{
			Name:    domainConfig.Name,
			IP:      ip,
			IPType:  domainConfig.IPType,
			Proxied: domainConfig.Proxied,
		}
		drs = append(drs, dr)
	}
	log.Infof("Start update dns: %+v\n", drs)
	if err := provider.UpdateDNS(domainConfig.Name, drs); err != nil {
		log.Error("Clean records failed: ", err)
	}
}
