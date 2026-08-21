package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL, HTTPAddr, AdminToken string
	WorkerConcurrency                 int
	AutoMigrate                       bool
}

func Load() Config {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: os.Getenv("HTTP_ADDR"), WorkerConcurrency: 2, AutoMigrate: os.Getenv("AUTO_MIGRATE") != "false"}
	c.AdminToken = os.Getenv("ADMIN_TOKEN")
	if c.DatabaseURL == "" {
		c.DatabaseURL = "postgres://postgres:postgres@localhost:5432/eventstore?sslmode=disable"
	}
	if c.HTTPAddr == "" {
		c.HTTPAddr = ":8080"
	}
	if value, err := strconv.Atoi(os.Getenv("WORKER_CONCURRENCY")); err == nil && value > 0 && value <= 32 {
		c.WorkerConcurrency = value
	}
	return c
}
