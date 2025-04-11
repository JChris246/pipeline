package main

import (
	"pipeline/data"

	"github.com/gin-gonic/gin"
)

func defineRoutes(router *gin.Engine /*logger *logrus.Logger*/) {
	const base = "/api"
	const register = base + "/register"

	// Define the route handlers

	// return the current version of the api
	router.GET(base+"/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": VERSION})
	})

	// register a pipeline with json definition
	router.POST(register+"/json", func(c *gin.Context) {
		var requestBody data.Pipeline
		if err := c.ShouldBind(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// register pipeline
		statusCode := 418 // 201
		c.JSON(statusCode, data.ApiErrorResponse{Message: ""})
		// c.JSON(statusCode, data.ApiErrorResponse{Message: "Pipeline registered"})
	})

	// register a pipeline with a file path
	router.POST(register+"/filepath", func(c *gin.Context) {
		var requestBody data.RegisterFilePath
		if err := c.ShouldBind(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// register pipeline
		statusCode := 418 // 201
		c.JSON(statusCode, data.ApiErrorResponse{Message: ""})
		// c.JSON(statusCode, data.ApiErrorResponse{Message: "Pipeline registered"})
	})
}
