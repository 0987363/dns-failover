package models

type PingResult struct {
	IP       string
	Latency  int
	LossRate float64
	Quality  float64
}
