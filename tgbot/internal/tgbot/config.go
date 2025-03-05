package tgbot

import (
	"time"
)

const(
	defaultTimeout = 10*time.Second
)

type Config struct {
	Token string 
	Admins []int 
	Timeout time.Duration
}

func NewConfig(token string, admins []int,timeout time.Duration) *Config {
	cfg:= &Config{Token: token, Admins: admins, Timeout: timeout}
	Validate(cfg)
	return cfg
}

func Validate(cfg *Config) {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.Token == "" {
		panic("token is empty")
	}
}