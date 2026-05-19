package config

import "os"

type Config struct {
	ServerPort string
	RedisAddr  string
	BaseDomain string
}

func Load() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "8090"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		BaseDomain: getEnv("BASE_DOMAIN", "localhost"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
