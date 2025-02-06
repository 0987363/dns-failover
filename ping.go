package main

import (
	"time"

	"github.com/0987363/dns-failover/models"
	probing "github.com/prometheus-community/pro-bing"
)

func CalculateQuality(latency int, lossRate float32) *models.PingResult {
	pr := &models.PingResult{
		Latency:  latency,
		LossRate: lossRate,
	}
	latencyScore := (float32(pr.Latency) / 500 * 100)
	packetScore := 100 - (pr.LossRate / 1 * 100)
	pr.Quality = 2.0 / (1.0/latencyScore + 1.0/packetScore)
	return pr
}

func Ping(address string, timeout time.Duration, sampling int) (*models.PingResult, error) {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		return nil, err
	}
	pinger.Count = sampling
	pinger.Interval = time.Millisecond * time.Duration(100)
	pinger.Timeout = time.Second * time.Duration(timeout)

	if err := pinger.Run(); err != nil {
		return nil, err
	}
	stats := pinger.Statistics()

	latency := stats.AvgRtt.Milliseconds()
	lossRate := 1.0 - float32(stats.PacketsRecv)/float32(stats.PacketsSent)
	tr := CalculateQuality(int(latency), lossRate)

	return tr, err
}
