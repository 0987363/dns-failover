package main

import (
	"math"
	"time"

	"github.com/0987363/dns-failover/models"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/sirupsen/logrus"
)

func CalculateQuality(latency int, lossRate float64) *models.PingResult {
	pr := &models.PingResult{
		Latency:  latency,
		LossRate: lossRate,
	}
	// 延迟得分计算（0ms得100分，每增加5ms扣1分）
	latencyScore := math.Max(0, 100.0-float64(latency)*0.2)
	// 丢包率得分计算（0%得100分，每增加1%扣10分）
	lossScore := math.Max(0, 100.0-lossRate/100*10)
	// 综合加权计算（延迟权重60%，丢包权重40%）
	pr.Quality = latencyScore*0.4 + lossScore*0.6
	return pr
}

func Ping(address string, timeout time.Duration, sampling int) (*models.PingResult, error) {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		return nil, err
	}
	pinger.Count = sampling
	pinger.Interval = time.Millisecond * time.Duration(100)
	pinger.Timeout = timeout
	if err := pinger.Run(); err != nil {
		return nil, err
	}

	stats := pinger.Statistics()
	logrus.Debugf("Ping stats: %+v\n", stats)

	latency := stats.AvgRtt.Milliseconds()
	lossRate := 1.0 - float64(stats.PacketsRecv)/float64(stats.PacketsSent)
	tr := CalculateQuality(int(latency), lossRate)

	return tr, err
}
