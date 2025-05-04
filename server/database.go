package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"pipeline/data"
	"pipeline/utils"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const PIPELINES_DIR = "pipelines"
const PIPELINE_RUNS = "pipeline_runs"
const REGISTERED_PIPELINES_FILE = "registered_pipelines.json"

func loadRegisteredPipelines(logger *logrus.Logger) map[string]data.RegisteredPipeline {
	utils.InitDataStoreDir(logger)

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINES_DIR, REGISTERED_PIPELINES_FILE)
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
	if !utils.InitDir(path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINES_DIR), logger) {
		logger.Error("Error creating pipelines directory: " + path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINES_DIR))
		return false
	}

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINES_DIR, REGISTERED_PIPELINES_FILE)

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

type ByDateString []fs.DirEntry

func (a ByDateString) Len() int      { return len(a) }
func (a ByDateString) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// maybe I should just save the file with the date number (unix timestamp) 🤦🏽‍♂️
func (a ByDateString) Less(i, j int) bool {
	if len(strings.Split(a[i].Name(), ".")) < 2 || len(strings.Split(a[j].Name(), ".")) < 2 {
		return false
	}

	// parsing this for every comparison seems wild
	var iPureName = strings.Split(a[i].Name(), ".")[0]
	var jPureName = strings.Split(a[j].Name(), ".")[0]
	var iTime, iErr = time.Parse(time.DateTime, strings.Replace(iPureName, "_", ":", 2))
	var jTime, jErr = time.Parse(time.DateTime, strings.Replace(jPureName, "_", ":", 2))

	if iErr != nil || jErr != nil {
		return false
	}
	return iTime.Before(jTime)
}

func loadPipelineRuns(logger *logrus.Logger, pipelineName string, limit int) []data.PipelineRun {
	utils.InitDataStoreDir(logger)

	var pipelineRunsDir = path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINE_RUNS, pipelineName)
	if _, err := os.Stat(pipelineRunsDir); os.IsNotExist(err) {
		logger.Warn("No runs for pipeline, " + pipelineName + ", returning empty array")
		return []data.PipelineRun{}
	}

	entries, err := os.ReadDir(pipelineRunsDir)
	if err != nil {
		logger.Error("Error reading "+pipelineRunsDir+" directory: ", err)
		return []data.PipelineRun{}
	}

	// TODO: filter out entries not ending in .json

	var pipelineRuns []data.PipelineRun
	var count = 0
	sort.Slice(entries, ByDateString(entries).Less)
	slices.Reverse(entries)

	for _, entry := range entries {
		count++
		if count > limit && limit != -1 {
			break
		}

		if len(strings.Split(entry.Name(), ".")) < 2 {
			continue
		}

		if strings.Split(entry.Name(), ".")[1] == "json" {
			var filename = path.Join(pipelineRunsDir, entry.Name())
			var fileData, err = os.ReadFile(filename)
			if err != nil {
				logger.Error("Error reading pipeline run file: " + filename)
				continue
			}

			var pipelineRun data.PipelineRun
			err = json.Unmarshal(fileData, &pipelineRun)
			if err != nil {
				logger.Error("Pipeline run file is corrupted, unable to parse JSON: " + filename)
				continue
			}

			pipelineRuns = append(pipelineRuns, pipelineRun)
		}
	}

	return pipelineRuns
}

func savePipelineRun(pipelineRun data.PipelineRun, logger *logrus.Logger) bool {
	utils.InitDataStoreDir(logger)

	if !utils.InitDir(path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINE_RUNS, pipelineRun.Name), logger) {
		logger.Error("Error creating pipeline runs directory: " + path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINE_RUNS, pipelineRun.Name))
		return false
	}

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), PIPELINE_RUNS, pipelineRun.Name, utils.GetCurrentTimeStamp(true)+".json")

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Error creating pipeline run file: " + err.Error())
		return false
	}

	err = json.NewEncoder(file).Encode(pipelineRun)
	if err != nil {
		logger.Error("Error writing to pipeline run file: " + err.Error())
		return false
	}

	err = file.Close()
	if err != nil {
		logger.Error("Error closing pipeline run file: " + err.Error())
		return false
	}

	return true
}
