package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
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

	viper.SetEnvKeyReplacer(nil)

	// Expand env vars before reading: supports both ${VAR} and ${VAR:-default}
	configContent := expandEnvVarsInFile("configs/config.yaml")
	if err := viper.ReadConfig(bytes.NewReader(configContent)); err != nil {
		panic(fmt.Sprintf("failed to read config: %v", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}

	return &cfg
}

// expandEnvVarsInFile reads a config file and expands ${VAR} and ${VAR:-default} placeholders
func expandEnvVarsInFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return []byte{}
	}
	return expandEnvVars(data)
}

// expandEnvVars expands ${VAR} and ${VAR:-default} patterns in a byte slice
func expandEnvVars(data []byte) []byte {
	// Match ${VAR} or ${VAR:-default}
	re := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)
	return re.ReplaceAllFunc(data, func(match []byte) []byte {
		matches := re.FindSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		key := string(matches[1])
		if val, ok := os.LookupEnv(key); ok {
			return []byte(val)
		}
		// No env var set: use default if provided, otherwise empty
		if len(matches) >= 3 {
			return matches[2]
		}
		return []byte{}
	})
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}
