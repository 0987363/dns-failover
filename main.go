package main

import (
	"fmt"
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

	// 如果没有设置 LOG_LEVEL，默认为 "info"
	if logLevel == "" {
		logLevel = "info"
	}

	// 将环境变量值转为小写，便于处理
	logLevel = strings.ToLower(logLevel)

	// 设置 logrus 的日志级别
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
		// 如果环境变量值不合法，默认设置为 info
		logrus.SetLevel(logrus.InfoLevel)
		fmt.Printf("Invalid log level: %s, defaulting to info\n", logLevel)
	}
}

func main() {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	provider := provider.Init(&config.ProviderConfig)
	for _, domainConfig := range config.Domains {
		go ScheduledTask(provider, domainConfig)
		time.Sleep(time.Second)
	}
	select {}
}
