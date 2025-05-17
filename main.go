package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/0987363/dns-failover/config"
	"github.com/0987363/dns-failover/provider"
	"github.com/sirupsen/logrus"
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevel = strings.ToLower(logLevel)

	switch logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})
	logrus.Info("Set log level: ", logLevel)
}

func main() {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}

	provider := provider.Init(&config.ProviderConfig)
	for _, domainConfig := range config.Domains {
		go ScheduledTask(provider, domainConfig)
		time.Sleep(time.Second)
	}
	select {}
}
