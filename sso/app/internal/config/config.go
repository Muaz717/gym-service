package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	DB       DBConfig      `yaml:"db"`
	GRPC     GRPCConfig    `yaml:"grpc"`
}

type DBConfig struct {
	Host       string `yaml:"host" env-required:"true"`
	DBPort     string `yaml:"port" env-required:"true"`
	Username   string `yaml:"username" env-required:"true"`
	DBName     string `yaml:"dbname" env-required:"true"`
	DBPassword string `yaml:"dbpassword" env-required:"true" env:"SSO_DB_PASSWORD"`
}
type GRPCConfig struct {
	Host    string        `yaml:"host" env-default:"localhost"`
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	cfgPath := fetchConfigPath()
	if cfgPath == "" {
		panic("config path is required, but it is empty")
	}

	return MustLoadByPath(cfgPath)
}

func MustLoadByPath(cfgPath string) *Config {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic("config file does not exist" + cfgPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
