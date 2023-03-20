package config

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
	"strings"

	"ferlab/envoy-transport-control-plane/logger"
)

type EtcdConfig struct {
	Prefix         string
	Endpoints      []string
	CaCert         string   `yaml:"ca_cert"`
	ClientCert     string   `yaml:"client_cert"`
	ClientKey      string   `yaml:"client_key"`
	Username       string
	Password       string
}

type ServerConfig struct {
	Port             int64
	BindIp           string        `yaml:"bind_ip"`
	MaxConnections   int64         `yaml:"max_connections"`
	KeepAliveTime    time.Duration `yaml:"keep_alive_time"`
	KeepAliveTimeout time.Duration `yaml:"keep_alive_timeout"`
	KeepAliveMinTime time.Duration `yaml:"keep_alive_min_time"`
}

type Config struct {
	Etcd EtcdConfig
	Server ServerConfig
	LogLevel string     `yaml:"log_level"`
}

func (c *Config) GetLogLevel() int64 {
	logLevel := strings.ToLower(c.LogLevel)
    switch logLevel {
    case "error":
        return logger.ERROR
    case "warning":
        return logger.WARN
    case "debug":
        return logger.DEBUG
	default:
		return logger.INFO
    }
}

func GetConfig(path string) (Config, error) {
	var c Config

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error reading the configuration file: %s", err.Error()))
	}

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error parsing the configuration file: %s", err.Error()))
	}

	return c, nil
}