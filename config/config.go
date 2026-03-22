package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const ConfigFilePath = "config/config.local.yaml"

type ConnConfig struct {
	Host           string `mapstructure:"Host"`
	Port           int    `mapstructure:"Port"`
	Database       string `mapstructure:"Database"`
	User           string `mapstructure:"User"`
	Password       string `mapstructure:"Password"`
	SSLMode        string `mapstructure:"SSLMode"`
	ConnectTimeout int    `mapstructure:"ConnectTimeout"`
}

func (c *ConnConfig) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=%d",
		c.User, c.Password, c.Host, c.Port,
		c.Database, c.SSLMode, c.ConnectTimeout,
	)
}

type PostgresPoolConfig struct {
	ConnConfig            ConnConfig    `mapstructure:"ConnConfig"`
	MaxConnLifetime       time.Duration `mapstructure:"MaxConnLifetime"`
	MaxConnLifetimeJitter time.Duration `mapstructure:"MaxConnLifetimeJitter"`
	MaxConnIdleTime       time.Duration `mapstructure:"MaxConnIdleTime"`
	MaxConns              int32         `mapstructure:"MaxConns"`
	MinConns              int32         `mapstructure:"MinConns"`
	HealthCheckPeriod     time.Duration `mapstructure:"HealthCheckPeriod"`
}

type PostgresConfig struct {
	Pool              PostgresPoolConfig `mapstructure:"pool"`
	ConnectRetries    int                `mapstructure:"ConnectRetries"`
	ConnectRetryDelay time.Duration      `mapstructure:"ConnectRetryDelay"`
	Schema            string             `mapstructure:"Schema"`
	MigrationsPath    string             `mapstructure:"MigrationsPath"`
}

type JWTConfig struct {
	SecretKey         string        `mapstructure:"secret_key"`
	AccessExpiration  time.Duration `mapstructure:"access_expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"ReadTimeout"`
	WriteTimeout    time.Duration `mapstructure:"WriteTimeout"`
	IdleTimeout     time.Duration `mapstructure:"IdleTimeout"`
	ShutDownTimeOut time.Duration `mapstructure:"ShutDownTimeOut"`
}

type LoggerConfig struct {
	Logger zap.Config `mapstructure:"logger"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
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
		log.Println("CONFIG_FILE_PATH not set, using default:", ConfigFilePath)
	}

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			zapLevelHook,
			stringToIntHook,
		),
		Result:  &cfg,
		TagName: "mapstructure",
	}

	dec, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return Config{}, fmt.Errorf("unable to create decoder: %w", err)
	}

	if err := dec.Decode(v.AllSettings()); err != nil {
		return Config{}, fmt.Errorf("unable to decode config: %w", err)
	}

	return cfg, nil
}
