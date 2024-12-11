package main

import (
	"github.com/go-redis/redis/v8"
	"sim-server/config"
	"sim-server/internal/handlers"
)

type Application struct {
	Config *config.Config
	Redis  *redis.Client
}

// handlers initializers
func (app *Application) simHandler() handlers.SimHandler {
	return handlers.SimHandler{}
}
