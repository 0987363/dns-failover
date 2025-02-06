package models

type PingResult struct {
	IP       string
	Latency  int
	LossRate float32
	Quality  float32
}
