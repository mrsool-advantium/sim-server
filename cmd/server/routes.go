package main

import (
	"github.com/gin-gonic/gin"
)

func (app *Application) registerRoutes(router *gin.Engine) {
	simHandler := app.simHandler()
	// Simulations
	simulation := router.Group("/simulation")
	{
		simulation.POST("/scenario", simHandler.SimulateScenario)
	}

}
