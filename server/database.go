package main

import (
	"encoding/json"
	"os"
	"path"
	"pipeline/data"
	"pipeline/utils"

	"github.com/sirupsen/logrus"
)

const REGISTERED_PIPELINES_FILE = "registered_pipelines.json"

func loadRegisteredPipelines(logger *logrus.Logger) map[string]data.RegisteredPipeline {
	utils.InitDataStoreDir(logger)

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), REGISTERED_PIPELINES_FILE)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Warn("No registered pipelines, returning empty map")
		return map[string]data.RegisteredPipeline{}
	}

	fileData, err := os.ReadFile(filename)
	if err != nil {
		logger.Error("Error reading registered pipelines file: " + filename)
		return map[string]data.RegisteredPipeline{}
	}

	var registeredPipelines map[string]data.RegisteredPipeline
	err = json.Unmarshal(fileData, &registeredPipelines)
	if err != nil {
		logger.Error("Registered pipelines file is corrupted, unable to parse JSON: " + filename)
		return map[string]data.RegisteredPipeline{}
	}

	return registeredPipelines
}

func saveRegisteredPipelines(pipelines map[string]data.RegisteredPipeline, logger *logrus.Logger) bool {
	utils.InitDataStoreDir(logger)

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), REGISTERED_PIPELINES_FILE)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("Error creating registered pipelines file: " + err.Error())
		return false
	}

	err = json.NewEncoder(file).Encode(pipelines)
	if err != nil {
		logger.Error("Error writing to registered pipelines file: " + err.Error())
		return false
	}

	err = file.Close()
	if err != nil {
		logger.Error("Error closing registered pipelines file: " + err.Error())
		return false
	}

	return true
}
