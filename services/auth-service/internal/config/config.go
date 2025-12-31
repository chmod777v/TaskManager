package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string  `yaml:"env" env:"APP_ENV" env-default:"development"`
	Service Service `yaml:"service"`
	Clients Clients `yaml:"clients"`
}

type Service struct {
	HTTPPort int    `yaml:"http_port" env:"SERVICE_HTTP_PORT" env-default:"8080"`
	GRPCPort int    `yaml:"grpc_port" env:"SERVICE_GRPC_PORT" env-default:"50050"`
	Host     string `yaml:"host" env:"SERVICE_HOST" env-default:"localhost"`
}

type Clients struct {
	TasksGRPC      string `yaml:"tasks_grpc" env:"CLIENTS_TASKS_GRPC" env-default:"localhost:50051"`
	AssignmentGRPC string `yaml:"assignment_grpc" env:"CLIENTS_ASSIGNMENT_GRPC" env-default:"localhost:50052"`
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
