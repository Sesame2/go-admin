package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"db_name"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

var (
	// GlobalConfig 全局实例
	GlobalConfig Config

	// 配置文件搜索路径
	configPaths = []string{
		".",
		"./configs",
		"../configs",
		"../../configs",
		"/etc/go-admin",
	}
)

// 加载配置
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 添加配置文件搜索路径
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 设置环境变量前缀，使用环境变量覆盖配置
	v.SetEnvPrefix("GO_ADMIN")
	v.AutomaticEnv()

	// 绑定特定环境变量
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.username", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.mode", "GIN_MODE")

	// 设置默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("jwt.expire", 24)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("未找到配置文件： %w", err)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置失败： %w", err)
	}

	// 设置全局配置
	GlobalConfig = *config

	return config, nil
}
