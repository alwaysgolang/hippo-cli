package config

import "time"

type Config struct {
	HTTP        HTTPConfig
	Application AppConfig
}

type HTTPConfig struct {
	Port int
	Mode string
}

type AppConfig struct {
	Mode              string
	TimeZone          *time.Location
	LogLevel          string
	ConsumeOnCallback bool
}
