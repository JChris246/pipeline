package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"pipeline/data"
	"pipeline/utils"
	"testing"
	"time"
)

var _ = os.Setenv("ENV", "test")
var _ = os.Setenv("DATA_STORE_DIR", "test_assets/data_store")
var _, testLogger = utils.SetupLogger("test.log")

// TODO: write better tests

func Test_loadPipelineRuns_ShouldReturnPipelineRunsInOrderOfDate(t *testing.T) {
	// arrange
	var pipelineRunPath = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline")
	utils.InitDir(pipelineRunPath, testLogger)

	const create = 10
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("2023-03-%02d %02d_00_%02d.json", i, i+2, i*2)
		var filename = path.Join(pipelineRunPath, name)
		var pipelineRun = data.PipelineRun{Name: fmt.Sprintf("test_pipeline_%d", i)}

		file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		json.NewEncoder(file).Encode(pipelineRun)
		file.Close()
	}

	// act
	var pipelineRuns = loadPipelineRuns(testLogger, "test_pipeline", -1)

	// assert
	utils.AssertEqual(t, create, len(pipelineRuns))
	for i := 0; i < create; i++ {
		utils.AssertStringEqual(t, fmt.Sprintf("test_pipeline_%d", create-i-1), pipelineRuns[i].Name)
	}

	// cleanup
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("2023-03-%02d %02d_00_%02d.json", i, i+2, i*2)
		var filename = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline", name)
		os.Remove(filename)
	}
}

func Test_loadPipelineRuns_ShouldReturnNPipelineRunsInOrderOfDate(t *testing.T) {
	// arrange
	var pipelineRunPath = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline")
	utils.InitDir(pipelineRunPath, testLogger)

	const create = 10
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("20%02d-03-%02d %02d_00_%02d.json", i*8, i, i+2, i*2)
		var filename = path.Join(pipelineRunPath, name)
		var pipelineRun = data.PipelineRun{Name: fmt.Sprintf("test_pipeline_%d", i), StartedAt: time.Now()}

		file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		json.NewEncoder(file).Encode(pipelineRun)
		file.Close()
	}

	// act
	const limit = 4
	var pipelineRuns = loadPipelineRuns(testLogger, "test_pipeline", limit)

	// assert
	utils.AssertEqual(t, limit, len(pipelineRuns))
	for i := 0; i < limit; i++ {
		utils.AssertStringEqual(t, fmt.Sprintf("test_pipeline_%d", create-i-1), pipelineRuns[i].Name)
	}

	// cleanup
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("20%02d-03-%02d %02d_00_%02d.json", i*8, i, i+2, i*2)
		var filename = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline", name)
		os.Remove(filename)
	}
}

func Test_loadPipelineRuns_ShouldReturnAllExistingPipelineRunsInOrderOfDateWhenLimitIsMoreThanExists(t *testing.T) {
	// arrange
	var pipelineRunPath = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline")
	utils.InitDir(pipelineRunPath, testLogger)

	const create = 5
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("20%02d-03-%02d %02d_00_%02d.json", i*8, i, i+2, i*2)
		var filename = path.Join(pipelineRunPath, name)
		var pipelineRun = data.PipelineRun{Name: fmt.Sprintf("test_pipeline_%d", i)}

		file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		json.NewEncoder(file).Encode(pipelineRun)
		file.Close()
	}

	// act
	const limit = 10
	var pipelineRuns = loadPipelineRuns(testLogger, "test_pipeline", limit)

	// assert
	utils.AssertEqual(t, create, len(pipelineRuns))
	for i := 0; i < create; i++ {
		utils.AssertStringEqual(t, fmt.Sprintf("test_pipeline_%d", create-i-1), pipelineRuns[i].Name)
	}

	// cleanup
	for i := 0; i < create; i++ {
		var name = fmt.Sprintf("20%02d-03-%02d %02d_00_%02d.json", i*8, i, i+2, i*2)
		var filename = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline_runs/test_pipeline", name)
		os.Remove(filename)
	}
}
