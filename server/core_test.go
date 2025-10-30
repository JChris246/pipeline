package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pipeline/data"
	"pipeline/utils"
	"runtime"
	"testing"
)

var _ = os.Setenv("ENV", "test")
var _ = os.Setenv("DATA_STORE_DIR", "test_assets/data_store")

const testPipeline = "test_assets/test_pipeline_%s.json"
const testSkipPipeline = "test_assets/test_pipeline_skip_%s.json"
const testNoParallelPipeline = "test_assets/test_pipeline_no_parallel_%s.json"
const testOverloadPipeline = "test_assets/test_pipeline_overload_%s.json"

func pipelineLoadHelper(pipelineFile string) data.Pipeline {
	var osSuffix = "linux"
	if runtime.GOOS == "windows" {
		osSuffix = "windows"
	}

	fileData, _ := os.ReadFile(fmt.Sprintf(pipelineFile, osSuffix))

	var pipeline data.Pipeline
	_ = json.Unmarshal(fileData, &pipeline)

	pipeline.Name = pipeline.Name + utils.GenerateId() // make the test runs unique
	return pipeline
}

func Test_runPipeline_ShouldReturnSuccessWhenAllTasksCompleted(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	// assert
	utils.AssertTrue(t, success)
	utils.AssertTrue(t, pipelineRun.Successful)
	for _, taskResponse := range pipelineRun.Stages {
		utils.AssertTrue(t, taskResponse.Successful)
	}

	// TODO: cleanup
}

func Test_runPipeline_ShouldReturnSuccessWhenAllTasksCompletedWhenPassingAPrecreatedPipelineRunObject(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testPipeline)
	var pipelineRun data.PipelineRun

	// act
	var success, _ = runPipeline(&pipeline, &pipelineRun, testLogger)

	// assert
	utils.AssertTrue(t, success)
	utils.AssertTrue(t, pipelineRun.Successful)
	for _, taskResponse := range pipelineRun.Stages {
		utils.AssertTrue(t, taskResponse.Successful)
	}

	// TODO: cleanup
}

// this test would probably be effective (/testable) if we were able to set the number of threads as the caller (coming soon?)
// technically the no parallel test, tests part of this
func Test_runPipeline_ShouldReturnSuccessWhenRunningParallelAndMoreTaskThanThreads(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testOverloadPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	// assert
	utils.AssertTrue(t, success)
	utils.AssertTrue(t, pipelineRun.Successful)
	for _, taskResponse := range pipelineRun.Stages {
		utils.AssertTrue(t, taskResponse.Successful)
	}

	// TODO: cleanup
}

func Test_runPipeline_ShouldRespectOrderAndDependencies(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	var taskMap map[string]data.TaskStatusResponse = make(map[string]data.TaskStatusResponse)
	for _, task := range pipelineRun.Stages {
		taskMap[task.TaskName] = task
	}

	// assert
	utils.AssertTrue(t, success)
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["initialize"].EndedAt.UnixMilli()), int(taskMap["build_frontend"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["initialize"].EndedAt.UnixMilli()), int(taskMap["build_backend"].StartedAt.UnixMilli()))

	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_frontend"].EndedAt.UnixMilli()), int(taskMap["run_tests"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_backend"].EndedAt.UnixMilli()), int(taskMap["run_tests"].StartedAt.UnixMilli()))

	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_frontend"].EndedAt.UnixMilli()), int(taskMap["security_scan"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_backend"].EndedAt.UnixMilli()), int(taskMap["security_scan"].StartedAt.UnixMilli()))

	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["run_tests"].EndedAt.UnixMilli()), int(taskMap["package"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["security_scan"].EndedAt.UnixMilli()), int(taskMap["package"].StartedAt.UnixMilli()))

	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["package"].EndedAt.UnixMilli()), int(taskMap["deploy_staging"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["deploy_staging"].EndedAt.UnixMilli()), int(taskMap["integration_tests"].StartedAt.UnixMilli()))

	// TODO: cleanup
}

func Test_runPipeline_ShouldReturnSuccessWhenAllTasksCompletedOrSetToSkip(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testSkipPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	// assert
	utils.AssertTrue(t, success)
	utils.AssertTrue(t, pipelineRun.Successful)
	for _, taskResponse := range pipelineRun.Stages {
		utils.AssertTrue(t, taskResponse.Successful)
	}

	// TODO: cleanup
}

func Test_runPipeline_ShouldSkipDependentTasksWhenDependencyTasksSkipped(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testSkipPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	var taskMap map[string]data.TaskStatusResponse = make(map[string]data.TaskStatusResponse)
	for _, task := range pipelineRun.Stages {
		taskMap[task.TaskName] = task
	}

	// assert
	utils.AssertTrue(t, success)

	// because security_scan is skipped all other dependent tasks (even transitively) should be skipped (but successful)
	utils.AssertTrue(t, taskMap["security_scan"].Skipped)
	utils.AssertTrue(t, taskMap["security_scan"].Successful)
	utils.AssertTrue(t, taskMap["security_scan"].StartedAt.IsZero())
	utils.AssertTrue(t, taskMap["package"].Skipped)
	utils.AssertTrue(t, taskMap["package"].Successful)
	utils.AssertTrue(t, taskMap["package"].StartedAt.IsZero())
	utils.AssertTrue(t, taskMap["deploy_staging"].Skipped)
	utils.AssertTrue(t, taskMap["deploy_staging"].Successful)
	utils.AssertTrue(t, taskMap["deploy_staging"].StartedAt.IsZero())
	utils.AssertTrue(t, taskMap["integration_tests"].Skipped)
	utils.AssertTrue(t, taskMap["integration_tests"].Successful)
	utils.AssertTrue(t, taskMap["integration_tests"].StartedAt.IsZero())

	// verify other tasks were not skipped
	utils.AssertFalse(t, taskMap["initialize"].Skipped)
	utils.AssertFalse(t, taskMap["initialize"].StartedAt.IsZero())
	utils.AssertFalse(t, taskMap["build_frontend"].Skipped)
	utils.AssertFalse(t, taskMap["build_frontend"].StartedAt.IsZero())
	utils.AssertFalse(t, taskMap["build_backend"].Skipped)
	utils.AssertFalse(t, taskMap["build_backend"].StartedAt.IsZero())
	utils.AssertFalse(t, taskMap["run_tests"].Skipped)
	utils.AssertFalse(t, taskMap["run_tests"].StartedAt.IsZero())

	// TODO: cleanup
}

func Test_runPipeline_ShouldRunSeriallyWhenParallelDisabled(t *testing.T) {
	t.Parallel()

	// arrange
	var pipeline data.Pipeline = pipelineLoadHelper(testNoParallelPipeline)

	// act
	var success, pipelineRun = runPipeline(&pipeline, nil, testLogger)

	var taskMap map[string]data.TaskStatusResponse = make(map[string]data.TaskStatusResponse)
	for _, task := range pipelineRun.Stages {
		taskMap[task.TaskName] = task
	}

	// assert
	utils.AssertTrue(t, success)

	// each task should run 1 after the other
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["initialize"].EndedAt.UnixMilli()), int(taskMap["build_frontend"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_frontend"].EndedAt.UnixMilli()), int(taskMap["build_backend"].StartedAt.UnixMilli()))
	utils.AssertGreaterThanOrEqualTo(t, int(taskMap["build_backend"].EndedAt.UnixMilli()), int(taskMap["run_tests"].StartedAt.UnixMilli()))

	// TODO: cleanup
}
