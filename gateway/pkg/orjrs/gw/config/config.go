package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expire time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AddConfigPath(".")

	// 自动展开环境变量 ${VAR} 或 ${VAR:-default}
	viper.SetEnvKeyReplacer(nil)

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read config: %v", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}

	return &cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}
