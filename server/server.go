package main

import (
	"encoding/json"
	"pipeline/data"
	"pipeline/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TODO: also store a map of active runs?
var Pipelines map[string]*data.PipelineItem = make(map[string]*data.PipelineItem, 20)

const NUM_LAST_RUNS = 10 // TODO: this could maybe be increased/decreased by ui

func initServer(logger *logrus.Logger) {
	logger.Info("Initializing server")

	var registeredPipelines = loadRegisteredPipelines(logger)
	for name, pipeline := range registeredPipelines {
		var runs = loadPipelineRuns(logger, name, NUM_LAST_RUNS)
		var lastRun int64 = 0
		// TODO: need to refactor to save pipeline run based on start time (instead of end time)
		// in order to use startAt be the determiner of last run
		if len(runs) != 0 {
			lastRun = runs[0].EndedAt.UnixMilli()
		}
		Pipelines[name] = &data.PipelineItem{
			Name:    pipeline.Name,
			Status:  data.PipelineStatus["IDLE"],
			LastRun: lastRun,
			Runs:    runs,
		}
	}
}

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
		registeredPipelineResponses := make([]data.RegisteredPipelineResponse, 0, len(registeredPipelines))
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

	// get pipeline runs
	router.GET(pipeline+"/:name/runs", func(c *gin.Context) {
		var runs, statusCode = getPipelineRuns(c.Param("name"), logger)
		c.JSON(statusCode, runs)
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

	var pipelineToValidate data.Pipeline
	b, _ := json.Marshal(pipelineRequest.PipelineDefinition)
	json.Unmarshal(b, &pipelineToValidate)

	var errors = utils.ValidatePipelineDefinition(&pipelineToValidate, nil, logger)
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

	Pipelines[pipelineRequest.PipelineDefinition.Name] = &data.PipelineItem{
		Name:    pipelineRequest.PipelineDefinition.Name,
		Status:  data.PipelineStatus["IDLE"],
		LastRun: 0,
		Runs:    make([]data.PipelineRun, 0),
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
			LastRun: Pipelines[name].LastRun,
			Status:  Pipelines[name].Status,
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
		LastRun:   Pipelines[pipeline.Name].LastRun,
		Status:    Pipelines[pipeline.Name].Status,
	}

	return &details, 200
}

func deletePipeline(name string, logger *logrus.Logger) (string, int) {
	if Pipelines[name].Status == data.PipelineStatus["RUNNING"] {
		logger.Warn("Pipeline " + name + " is running, cannot delete")
		return "Pipeline is running, cannot delete", 409
	}

	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	utils.DeleteFile(registeredPipelines[name].Path, logger)
	utils.DeleteFile(registeredPipelines[name].VariablesFile, logger)
	delete(registeredPipelines, name)

	saveRegisteredPipelines(registeredPipelines, logger)
	delete(Pipelines, name)

	return "Pipeline deleted", 200
}

func editPipeline(name string, pipelineRequest *data.EditPipelineRequest, logger *logrus.Logger) (string, int) {
	if Pipelines[name].Status == data.PipelineStatus["RUNNING"] {
		logger.Warn("Pipeline " + name + " is running, cannot edit")
		return "Pipeline is running, cannot edit", 409
	}

	var registeredPipelines = loadRegisteredPipelines(logger)

	if _, exists := registeredPipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	// check if request is to change the pipeline to one that already exists
	if _, exists := registeredPipelines[pipelineRequest.Name]; exists && pipelineRequest.Name != name {
		logger.Warn("Cannot change pipeline name: '" + pipelineRequest.Name + "' already exists")
		return "Cannot change pipeline name: '" + pipelineRequest.Name + "' already exists", 409
	}

	var editPipeline = data.Pipeline{
		Name:     pipelineRequest.Name,
		Stages:   pipelineRequest.Stages,
		Parallel: pipelineRequest.Parallel,
	}

	var stages []data.Stage
	b, _ := json.Marshal(pipelineRequest.Stages)
	json.Unmarshal(b, &stages)

	var editPipelineToValidate = data.Pipeline{
		Name:     pipelineRequest.Name,
		Stages:   stages,
		Parallel: pipelineRequest.Parallel,
	}

	var errors = utils.ValidatePipelineDefinition(&editPipelineToValidate, &pipelineRequest.Variables, logger)
	if len(errors) > 0 {
		logger.Warn("Invalid pipeline definition: " + strings.Join(errors, "\n"))
		return "Invalid pipeline definition: " + strings.Join(errors, "\n"), 400
	}

	if len(pipelineRequest.Variables) > 0 && registeredPipelines[name].VariablesFile == "" {
		var varsFile = utils.CreateVariableFile(pipelineRequest.Variables, logger)
		if varsFile == "" {
			logger.Error("Error creating variable file: " + varsFile)
			return "Error creating variable file", 500
		}
		registeredPipelines[name] = data.RegisteredPipeline{
			Name:          registeredPipelines[name].Name,
			VariablesFile: varsFile,
			Path:          registeredPipelines[name].Path,
		}
	} else if len(pipelineRequest.Variables) > 0 {
		// if this fails it has the potential to break the pipeline because of missing variables
		var error = utils.SaveVariableFile(pipelineRequest.Variables, registeredPipelines[name].VariablesFile, logger)
		if error != "" {
			return "Error saving variable file", 500
		}
	}
	editPipeline.VariableFile = registeredPipelines[name].VariablesFile

	var filename = utils.SavePipelineDefinition(&editPipeline, logger)
	if filename == "" {
		return "Error saving pipeline definition", 500
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

		Pipelines[pipelineRequest.Name] = &data.PipelineItem{
			Name:    pipelineRequest.Name,
			Status:  Pipelines[name].Status,
			LastRun: Pipelines[name].LastRun,
			Runs:    Pipelines[name].Runs,
		}
		delete(Pipelines, name)
	}

	return "Pipeline updated", 200
}

func launchPipeline(name string, logger *logrus.Logger) (string, int) {
	if _, exists := Pipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist, can't launch")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	if Pipelines[name].Status == data.PipelineStatus["RUNNING"] {
		logger.Warn("Pipeline " + name + " is already running, will not start new run")
		return "Pipeline is not running, will not start new run", 409
	}

	logger.Info("Launching pipeline " + name)
	return name, 418
}

func cancelPipeline(name string, logger *logrus.Logger) (string, int) {
	if _, exists := Pipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist, can't cancel")
		return "Pipeline with name '" + name + "' does not exist", 404
	}

	if Pipelines[name].Status != data.PipelineStatus["RUNNING"] {
		logger.Warn("Pipeline " + name + " is not running, cannot cancel")
		return "Pipeline is not running, cannot cancel", 409
	}

	logger.Info("Cancelling pipeline " + name)
	return name, 418
}

func getPipelineRuns(name string, logger *logrus.Logger) ([]data.PipelineRun, int) {
	if _, exists := Pipelines[name]; !exists {
		logger.Warn("Pipeline with name '" + name + "' does not exist, can't get runs")
		return []data.PipelineRun{}, 404
	}

	logger.Info("Getting pipeline runs for " + name)

	// TODO: pull from Pipelines n number of pipeline runs (and if a run is in the active list return?)
	return []data.PipelineRun{}, 200
}
