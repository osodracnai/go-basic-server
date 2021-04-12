package cmd

import (
	"github.com/Shopify/sarama"
)

type (
	ConfigServerCmd struct {
		Listen string `mapstructure:"listen"`
		Debug  bool   `mapstructure:"debug"`
	}
	ConfigConsumerCmd struct {
		Jaeger Jaeger
		Kafka  Kafka
		Topics []string
	}
	Jaeger struct {
		Name     string
		LogError bool `mapstructure:"log-error"`
		LogInfo  bool `mapstructure:"log-info"`
	}
	Kafka struct {
		Brokers  []string
		GroupID  string `mapstructure:"group-id"`
		Version  string
		Client   sarama.Client
		Consumer sarama.ConsumerGroup
		Producer sarama.SyncProducer
	}
)
