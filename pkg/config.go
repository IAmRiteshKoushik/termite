package pkg

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

type Config struct {
	LogEnv      string `mapstructure:"env"`
	RabbitMQURL string `mapstructure:"rabbitmq_url"`
	WoC         struct {
		WebhookURL string `mapstructure:"webhook_url"`
		QueueName  string `mapstructure:"queue_name"`
	} `mapstructure:"woc"`
	AIVerse struct {
		WebhookURL string `mapstructure:"webhook_url"`
		QueueName  string `mapstructure:"queue_name"`
	} `mapstructure:"aiverse"`
}

var AppConfig *Config

func LoadConfig() error {
	viper.SetConfigName("env")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateURL(config.RabbitMQURL); err != nil {
		return fmt.Errorf("invalid RabbitMQ URL: %w", err)
	}
	if err := validateURL(config.WoC.WebhookURL); err != nil {
		return fmt.Errorf("invalid WoC webhook URL: %w", err)
	}
	if err := validateURL(config.AIVerse.WebhookURL); err != nil {
		return fmt.Errorf("invalid AIVerse webhook URL: %w", err)
	}

	AppConfig = &config
	return nil
}

func validateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}
	return nil
}
