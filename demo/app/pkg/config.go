package pkg

import "time"

type Config struct {
	AccessTokenExp time.Duration
	RefreshTokenExp time.Duration
}

func NewConfig() *Config {
	return &Config{
		AccessTokenExp:  15 * time.Minute,
		RefreshTokenExp: 30 * 24 * time.Hour,
	}
}
