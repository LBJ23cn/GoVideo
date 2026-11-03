package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	RabbitMQ      RabbitMQConfig      `yaml:"rabbitmq"`
	JWT           JWTConfig           `yaml:"jwt"`
	OSS           OSSConfig           `yaml:"oss"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type JWTConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type OSSConfig struct {
	Bucket          string `yaml:"bucket"`
	Region          string `yaml:"region"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	PublicBaseURL   string `yaml:"public_base_url"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
}

type ObservabilityConfig struct {
	Pprof PprofConfig `yaml:"pprof"`
}

type PprofConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ApiAddr    string `yaml:"api_addr"`
	WorkerAddr string `yaml:"worker_addr"`
}

// Load 加载文件配置
func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}
	// 加载环境配置 允许环境配置覆盖文件配置
	ApplyEnvOverrides(&config)
	return config, nil
}

// ApplyEnvOverrides 加载环境配置 可覆盖文件配置
func ApplyEnvOverrides(config *Config) {
	if config == nil {
		return
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Server.Port = port
		}
	}
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		config.Database.Host = v
	}
	if v := os.Getenv("MYSQL_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Database.Port = port
		}
	}
	if v := os.Getenv("MYSQL_USER"); v != "" {
		config.Database.User = v
	}
	if v := os.Getenv("MYSQL_ROOT_PASSWORD"); v != "" {
		config.Database.Password = v
	}
	if v := os.Getenv("MYSQL_PASSWORD"); v != "" {
		config.Database.Password = v
	}
	if v := os.Getenv("MYSQL_DATABASE"); v != "" {
		config.Database.DBName = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		config.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		config.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			config.Redis.DB = db
		}
	}
	if v := os.Getenv("RABBITMQ_HOST"); v != "" {
		config.RabbitMQ.Host = v
	}
	if v := os.Getenv("RABBITMQ_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.RabbitMQ.Port = port
		}
	}
	if v := os.Getenv("RABBITMQ_USER"); v != "" {
		config.RabbitMQ.Username = v
	}
	if v := os.Getenv("RABBITMQ_PASS"); v != "" {
		config.RabbitMQ.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		config.JWT.JWTSecret = v
	}
	if v := os.Getenv("OSS_BUCKET"); v != "" {
		config.OSS.Bucket = v
	}
	if v := os.Getenv("OSS_REGION"); v != "" {
		config.OSS.Region = v
	}
	if v := os.Getenv("OSS_ENDPOINT"); v != "" {
		config.OSS.Endpoint = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_ID"); v != "" {
		config.OSS.AccessKeyID = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_SECRET"); v != "" {
		config.OSS.AccessKeySecret = v
	}
	if v := os.Getenv("OSS_PUBLIC_BASE_URL"); v != "" {
		config.OSS.PublicBaseURL = v
	}
	if v := os.Getenv("OSS_TIMEOUT_SECONDS"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil {
			config.OSS.TimeoutSeconds = timeout
		}
	}
}

// TODO 需不需要写加载默认配置？
