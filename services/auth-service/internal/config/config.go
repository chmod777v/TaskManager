package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"development"`
	Server   Server   `yaml:"server"`
	Redis    Redis    `yaml:"redis"`
	Services Services `yaml:"services"`
}

type Server struct {
	GrpcPort int    `yaml:"grpc_port" env-default:"50050"`
	Host     string `yaml:"host" env-default:"localhost"`
}

type Redis struct {
	Port     int    `yaml:"port" env-default:"6379"`
	Host     string `yaml:"host" env-default:"localhost"`
	Password string `yaml:"password" env-default:""`
}

type Services struct {
	Users ServiceConfig `yaml:"users"`
}

type ServiceConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	GrpcPort int    `yaml:"grpc_port" env-default:"50051"`
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
