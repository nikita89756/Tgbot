package config

import (
	"flag"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct{
	App AppConfig `yaml:"app"`
	Bot BotConfig `yaml:"bot"`
	Storage StorageConfig `yaml:"storage"`
	Server Server `yaml:"server"`
}

type AppConfig struct{
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	LogLevel string `yaml:"logLevel"`
}

type BotConfig struct{
	Token string `yaml:"token"`
	Admins []int `yaml:"admins"`
	Timeout time.Duration `yaml:"timeout"`
}

type Server struct{
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	JwtToken string `yaml:"jwt_token"`
	TokenTTL time.Duration `yaml:"tokenTTL"`
}

type StorageConfig struct{
	ConnectionString string `yaml:"connectionString"`
	Type string `yaml:"type"`
}
var(
	once sync.Once

	config *Config
)

func NewConfig() *Config {
    once.Do(func() {
        var configPath string
        flag.StringVar(&configPath, "config", "", "path to config file")
        flag.Parse()

        if configPath == "" {
            panic("config path is empty")
        }

        config = &Config{}

        if err := cleanenv.ReadConfig(configPath, config); err != nil {
            panic(err)
        }
    })
    return config
}
