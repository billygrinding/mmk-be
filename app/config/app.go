package config

import "time"

type App struct {
	ServiceName      string        `envconfig:"SERVICE_NAME" default:"money-manager-be"`
	Env              string        `envconfig:"APP_ENV" default:"local"`
	ContextTimeout   time.Duration `envconfig:"CONTEXT_TIMEOUT" default:"10s"`
	DefaultCacheTTL  time.Duration `envconfig:"DEFAULT_CACHE_TTL" default:"3m"`
	DBConnectTimeout time.Duration `envconfig:"DB_CONNECT_TIMEOUT" default:"1m"`
	Port             string        `envconfig:"PORT" default:"2000"`
}
