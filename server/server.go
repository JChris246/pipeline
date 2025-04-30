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
		registeredPipelineResponses := make([]data.RegisteredPipelineResponse, 0)
		transformRegisteredPipelines(&registeredPipelines, &registeredPipelineResponses)
		c.JSON(200, registeredPipelineResponses)
	})

	// return the details of a pipeline (including run status)
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

	// delete a pipeline
	router.DELETE(register+"/:name", func(c *gin.Context) {
		var msg, statusCode = deletePipeline(c.Param("name"), logger)
		c.JSON(statusCode, data.ApiErrorResponse{Message: msg})
	})

	// edit a pipeline
	router.PATCH(pipeline+"/:name", func(c *gin.Context) {
		var requestBody data.EditPipelineRequest
		if err := c.ShouldBind(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var msg, statusCode = editPipeline(c.Param("name"), &requestBody, logger)
		c.JSON(statusCode, data.ApiErrorResponse{Message: msg})
	})

	// start a pipeline run
	router.POST(pipeline+"/:name", func(c *gin.Context) {
		var msg, statusCode = launchPipeline(c.Param("name"), logger)
		c.JSON(statusCode, data.ApiErrorResponse{Message: msg})
	})

	// cancel a pipeline run
	router.DELETE(pipeline+"/:name", func(c *gin.Context) {
		var msg, statusCode = cancelPipeline(c.Param("name"), logger)
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

	var errors = utils.ValidatePipelineDefinition(&pipelineRequest.PipelineDefinition, nil, logger)
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

	// this also has the potential to leave data in a weird state if this fails
	if !saveRegisteredPipelines(registeredPipelines, logger) {
		return "Error saving registered pipelines", 500
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

	var errors = utils.ValidatePipelineDefinition(newPipeline, nil, logger)
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

func deletePipeline(name string, logger *logrus.Logger) (string, int) {
	// TODO: don't allow deleting a pipeline if it is running
	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	utils.DeleteFile(registeredPipelines[name].Path, logger)
	utils.DeleteFile(registeredPipelines[name].VariablesFile, logger)
	delete(registeredPipelines, name)

	saveRegisteredPipelines(registeredPipelines, logger)

	return "Pipeline deleted", 204
}

func editPipeline(name string, pipelineRequest *data.EditPipelineRequest, logger *logrus.Logger) (string, int) {
	// TODO: don't allow editing a pipeline if it is running
	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	// check if request is to change the pipeline to one that already exists
	if _, exists := registeredPipelines[pipelineRequest.Name]; exists {
		logger.Warn("Cannot change pipeline name: '" + pipelineRequest.Name + "' already exists")
		return "Cannot change pipeline name: '" + pipelineRequest.Name + "' already exists", 409
	}

	var editPipeline = data.Pipeline{
		Name:     pipelineRequest.Name,
		Stages:   pipelineRequest.Stages,
		Parallel: pipelineRequest.Parallel,
	}

	var errors = utils.ValidatePipelineDefinition(&editPipeline, &pipelineRequest.Variables, logger)
	if len(errors) > 0 {
		logger.Warn("Invalid pipeline definition: " + strings.Join(errors, "\n"))
		return "Invalid pipeline definition: " + strings.Join(errors, "\n"), 400
	}

	editPipeline.VariableFile = registeredPipelines[name].VariablesFile

	var filename = utils.SavePipelineDefinition(&editPipeline, logger)
	if filename == "" {
		return "Error saving pipeline definition", 500
	}

	// if this fails it has the potential to break the pipeline because of missing variables
	var error = utils.SaveVariableFile(pipelineRequest.Variables, registeredPipelines[name].VariablesFile, logger)
	if error != "" {
		return "Error saving variable file", 500
	}

	// if the name has changed, update index in registeredPipelines
	if pipelineRequest.Name != name {
		delete(registeredPipelines, name)
		registeredPipelines[pipelineRequest.Name] = data.RegisteredPipeline{
			Name:          pipelineRequest.Name,
			VariablesFile: editPipeline.VariableFile,
			Path:          filename,
		}

		// this also has the potential to leave data in a weird state if this fails
		if !saveRegisteredPipelines(registeredPipelines, logger) {
			return "Error saving registered pipelines", 500
		}
	}

	return "Pipeline updated", 200
}

func launchPipeline(name string, logger *logrus.Logger) (string, int) {
	// TODO: don't allow launching a new pipeline run if it is already running
	logger.Info("Launching pipeline " + name)
	return name, 418
}

func cancelPipeline(name string, logger *logrus.Logger) (string, int) {
	logger.Info("Cancelling pipeline " + name)
	return name, 418
}
