package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/opensearch-project/opensearch-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const ConfigFilePath = "config/config.local.yaml"

type JWTConfig struct {
	SecKey     string        `mapstructure:"seckey"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type LoggerConfig struct {
	Logger zap.Config `mapstructure:"logger"`
}

func (l *LoggerConfig) Build() (*zap.Logger, error) {
	return l.Logger.Build()
}

type SearchConnect struct {
	Retries int           `mapstructure:"retries"`
	Delay   time.Duration `mapstructure:"delay"`
}

type SearchConfig struct {
	Connect SearchConnect     `mapstructure:"connect"`
	Client  opensearch.Config `mapstructure:"client"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"ReadTimeout"`
	WriteTimeout    time.Duration `mapstructure:"WriteTimeout"`
	IdleTimeout     time.Duration `mapstructure:"IdleTimeout"`
	ShutDownTimeOut time.Duration `mapstructure:"ShutDownTimeOut"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Search SearchConfig `mapstructure:"opensearch"`
	Logger LoggerConfig `yaml:"logger"`
	JWT    JWTConfig    `yaml:"jwt"`
}

func LoadConfig() (Config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("APP")

	if cfgFile := os.Getenv("CONFIG_FILE_PATH"); cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigFile(ConfigFilePath)
		log.Println("failed to read CONFIG_FILE_PATH, using default path")
	}

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("FATAL: error reading config file: %w", err)
	}

	var cfg Config

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(), // чтобы мапить durations
			zapLevelHook,    // хук для AtomicLevel
			stringToIntHook, // хук чтобы из .env доставать порт
		),
		Result:  &cfg,
		TagName: "mapstructure",
	}

	dec, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return Config{}, fmt.Errorf("FATAL:unable to create new decoder: %w", err)
	}

	if err := dec.Decode(v.AllSettings()); err != nil {
		return Config{}, fmt.Errorf("FATAL:unable to decode into struct: %w", err)
	}

	return cfg, nil
}
