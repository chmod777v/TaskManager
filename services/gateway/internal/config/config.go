package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	Server   Server   `yaml:"server"`
	Services Services `yaml:"services"`
}

type Server struct {
	HttpPort    int           `yaml:"http_port" env-default:"8080"`
	Host        string        `yaml:"host" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"30s"`
}

type Services struct {
	Auth       ServiceConfig `yaml:"auth"`
	Users      ServiceConfig `yaml:"users"`
	Tasks      ServiceConfig `yaml:"tasks"`
	Assignment ServiceConfig `yaml:"assignment"`
}

type ServiceConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	GrpcPort int    `yaml:"grpc_port"`
}

func LoadConfig() *Config {
	var path string
	flag.StringVar(&path, "config", "", "path") //"config" - имя флага (--config) "path" - описание для справки
	flag.Parse()
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		panic("CONFIG_PATH is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) { //os.Stat- - Проверка существует ли файл, os.IsNotExist-если нет то
		panic("Config file does not exist: " + path)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("Failet to read config" + err.Error())
	}
	return &cfg
}
