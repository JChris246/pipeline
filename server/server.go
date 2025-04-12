package main

import (
	"pipeline/data"
	"pipeline/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func defineRoutes(router *gin.Engine, logger *logrus.Logger) {
	const base = "/api"
	const pipeline = base + "/pipelines"
	const register = pipeline + "/register"

	// Define the route handlers

	// return the current version of the api
	router.GET(base+"/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": VERSION})
	})

	// return the list of registered pipelines
	router.GET(pipeline, func(c *gin.Context) {
		var registeredPipelines = loadRegisteredPipelines(logger)
		registeredPipelineResponses := make([]data.RegisteredPipelineResponse, len(registeredPipelines))
		transformRegisteredPipelines(&registeredPipelines, &registeredPipelineResponses)
		c.JSON(200, registeredPipelineResponses)
	})

	router.GET(pipeline+"/:name", func(c *gin.Context) {
		var details, statusCode = getPipelineDetails(c.Param("name"), logger)
		c.JSON(statusCode, details)
	})

	// register a pipeline with json definition
	router.POST(register+"/json", func(c *gin.Context) {
		var requestBody data.RegisterPipelineRequest
		if err := c.ShouldBind(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// upload and register pipeline
		var msg, statusCode = uploadPipelineDefinition(&requestBody, logger)
		c.JSON(statusCode, data.ApiErrorResponse{Message: msg})
	})

	// register a pipeline with a file path
	router.POST(register+"/filepath", func(c *gin.Context) {
		var requestBody data.RegisterFilePath
		if err := c.ShouldBind(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// register pipeline
		var msg, statusCode = registerPipelineFromFile(&requestBody, logger)
		c.JSON(statusCode, data.ApiErrorResponse{Message: msg})
	})
}

func uploadPipelineDefinition(pipelineRequest *data.RegisterPipelineRequest, logger *logrus.Logger) (string, int) {
	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[pipelineRequest.PipelineDefinition.Name]; exists {
		return "Pipeline with name '" + pipelineRequest.PipelineDefinition.Name + "' already exists", 409
	}

	var varsFile = ""
	if len(pipelineRequest.Variables) > 0 {
		var varsFile = utils.CreateVariableFile(pipelineRequest.Variables, logger)
		if varsFile == "" {
			logger.Error("Error creating variable file: " + varsFile)
			return "Error creating variable file", 500
		}
		pipelineRequest.PipelineDefinition.VariableFile = varsFile
	}

	var errors = utils.ValidatePipelineDefinition(&pipelineRequest.PipelineDefinition, logger)
	if len(errors) > 0 {
		logger.Warn("Invalid pipeline definition: " + strings.Join(errors, "\n"))
		utils.DeleteFile(varsFile, logger)
		return "Invalid pipeline definition: " + strings.Join(errors, "\n"), 400
	}

	var filename = utils.SavePipelineDefinition(&pipelineRequest.PipelineDefinition, logger)
	if filename == "" {
		logger.Error("Error saving pipeline definition")
		utils.DeleteFile(varsFile, logger)
		return "Error saving pipeline definition", 500
	}

	registeredPipelines[pipelineRequest.PipelineDefinition.Name] = data.RegisteredPipeline{
		Name:          pipelineRequest.PipelineDefinition.Name,
		VariablesFile: pipelineRequest.PipelineDefinition.VariableFile, // this is probably not needed
		Path:          filename,
	}

	return "Pipeline registered", 201
}

func registerPipelineFromFile(pipelineRequest *data.RegisterFilePath, logger *logrus.Logger) (string, int) {
	var registeredPipelines = loadRegisteredPipelines(logger)

	if pipelineRequest.DefinitionFilePath == "" {
		return "Missing pipeline definition path", 400
	}

	var newPipeline = utils.LoadDefinition(pipelineRequest.DefinitionFilePath, logger)
	if newPipeline == nil {
		logger.Warn("Invalid pipeline definition file")
		return "Invalid pipeline definition file", 400
	}

	if _, exists := registeredPipelines[newPipeline.Name]; exists {
		logger.Warn("Pipeline with name '" + newPipeline.Name + "' already exists")
		return "Pipeline with name '" + newPipeline.Name + "' already exists", 409
	}

	var errors = utils.ValidatePipelineDefinition(newPipeline, logger)
	if len(errors) > 0 {
		logger.Warn("Invalid pipeline definition: " + strings.Join(errors, "\n"))
		return "Invalid pipeline definition: " + strings.Join(errors, "\n"), 400
	}

	registeredPipelines[newPipeline.Name] = data.RegisteredPipeline{
		Name:          newPipeline.Name,
		VariablesFile: newPipeline.VariableFile, // this is probably not needed
		Path:          pipelineRequest.DefinitionFilePath,
	}

	return "Pipeline registered", 201
}

func transformRegisteredPipelines(registeredPipelines *map[string]data.RegisteredPipeline, registeredPipelineResponses *[]data.RegisteredPipelineResponse) {
	for name := range *registeredPipelines {
		(*registeredPipelineResponses) = append(*registeredPipelineResponses, data.RegisteredPipelineResponse{
			Name:    name,
			LastRun: 0,                           // TODO: load from in memory database
			Status:  data.PipelineStatus["IDLE"], // TODO: load from in memory database
		})
	}
}

func getPipelineDetails(name string, logger *logrus.Logger) (*data.RegisteredPipelineDetails, int) {
	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist")
		return nil, 404
	}

	var pipeline = utils.LoadDefinition(registeredPipelines[name].Path, logger)
	if pipeline == nil {
		// a pipeline is registered without actually existing
		logger.Error("Couldn't find pipeline with name '" + name + "'")
		return nil, 500
	}

	var variables = make(map[string]string)
	if pipeline.VariableFile != "" {
		variables = utils.LoadPipelineVars(pipeline.VariableFile, logger)
	}

	var details = data.RegisteredPipelineDetails{
		Name:      pipeline.Name,
		Stages:    pipeline.Stages,
		Parallel:  pipeline.Parallel,
		Variables: variables,
		LastRun:   0,                           // TODO: load from in memory database
		Status:    data.PipelineStatus["IDLE"], // TODO: load from in memory database
	}

	return &details, 200
}
