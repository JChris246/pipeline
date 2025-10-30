package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"pipeline/data"
	"pipeline/utils"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func runTask(stage data.Stage, pipelineName string) (bool, string) {
	cmd := exec.Command(stage.Task, stage.Args...)
	cmd.Dir = stage.Pwd
	cmd.Env = stage.Env

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return false, err.Error()
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return false, err.Error()
	}

	err = cmd.Start()
	if err != nil {
		return false, err.Error()
	}

	if pipelineName == "" {
		pipelineName = "pipeline"
	}

	var outputLogName = utils.CreateOutputLogName(pipelineName, stage.Name, false)
	logFile, err := os.OpenFile(outputLogName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err.Error()
	}
	defer logFile.Close()

	// Do we really want a separate file for the error logs?
	// var errorLogName = utils.CreateOutputLogName(pipelineName, stage.Name, true)
	// errorLogFile, err := os.OpenFile(errorLogName, os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return false, err.Error()
	// }
	// defer errorLogFile.Close()

	logReaderWg := sync.WaitGroup{}
	logReaderWg.Add(2)

	go func() {
		defer logReaderWg.Done()

		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			logFile.WriteString(line + "\n")
		}

		// if err := scanner.Err(); err != nil {
		// 	// logger.Error("Failed to read stdout: " + err.Error())
		// }
	}()

	go func() {
		defer logReaderWg.Done()

		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			logFile.WriteString(line + "\n")
		}

		// if err := scanner.Err(); err != nil {
		// 	// logger.Error("Failed to read stderr: " + err.Error())
		// }
	}()

	logReaderWg.Wait()
	err = cmd.Wait()
	if err != nil {
		return false, err.Error()
	}

	return true, ""
}

// for api server this will need to run on a separate thread?
// these logs are useful in headless mode, but in server mode they will probably be a log of noise.
// consider disabling them when running in server mode?
func runPipeline(pipeline *data.Pipeline, pipelineRun *data.PipelineRun, logger *logrus.Logger) (bool, data.PipelineRun) {
	var threads = 1
	if pipeline.Parallel {
		threads = runtime.NumCPU() / 2 // should this be configurable?
	}
	logger.Debug("Running pipeline " + pipeline.Name + " with " + fmt.Sprint(threads) + " thread(s)")

	if pipelineRun == nil {
		pipelineRun = &data.PipelineRun{
			Name:       pipeline.Name,
			StartedAt:  time.Now(),
			Successful: true,
			Stages:     make([]data.TaskStatusResponse, 0, len(pipeline.Stages)),
		}
	} else {
		pipelineRun.Name = pipeline.Name
		pipelineRun.StartedAt = time.Now()
		pipelineRun.Successful = true
		pipelineRun.Stages = make([]data.TaskStatusResponse, 0, len(pipeline.Stages))
	}

	// could have replaced these with the mutex, but I liked the channel approach I originally had for collecting task responses at completion
	taskResponses := make(map[string]data.TaskStatusResponse, len(pipeline.Stages))
	taskStatusBuffer := make(chan data.TaskStatusResponse, threads)

	var taskWg sync.WaitGroup
	var activeThreads = 0
	var pipelineMutex sync.Mutex // Mutex to protect pipelineRun updates

	// Helper function to safely update pipelineRun with completed task
	updatePipelineRun := func(taskResponse data.TaskStatusResponse) {
		pipelineMutex.Lock()
		defer pipelineMutex.Unlock()

		// Find and update existing stage or append new one
		updated := false
		for i, stage := range pipelineRun.Stages {
			if stage.TaskName == taskResponse.TaskName {
				pipelineRun.Stages[i] = taskResponse
				updated = true
				break
			}
		}

		// not updated, must be a new entry
		if !updated {
			pipelineRun.Stages = append(pipelineRun.Stages, taskResponse)
		}
	}

	collect := func(reInit bool) {
		taskWg.Wait()
		close(taskStatusBuffer)
		for i := range taskStatusBuffer {
			taskResponses[i.TaskName] = i
			updatePipelineRun(i)
		}
		if reInit {
			taskStatusBuffer = make(chan data.TaskStatusResponse, threads)
			activeThreads = 0
		}
	}

	for _, stage := range pipeline.Stages {
		// making tasks wait on their dependencies. this why it is critical to define
		// the order of tasks carefully, because this will block all subsequent tasks
		// this is by design to keep it simple for now
		if len(stage.DependsOn) > 0 {
			logger.Debug("Checking dependencies for: " + stage.Name)
			var dependenciesFailed = false
			for _, dependency := range stage.DependsOn {
				if _, ok := taskResponses[dependency]; !ok {
					logger.Info("Waiting for dependency: '" + dependency + "' for stage: " + stage.Name)
					collect(true) // wait for dependency to finish
					logger.Info("Dependency finished: '" + dependency + "' for stage: " + stage.Name)
				}

				if !taskResponses[dependency].Successful {
					logger.Warn("Dependency failed: '" + dependency + "' for stage: " + stage.Name + " skipping this stage")
					// skip this stage as the dependency failed, also mark this stage as failed
					taskResponses[stage.Name] = data.TaskStatusResponse{TaskName: stage.Name, Successful: false, Skipped: true}
					dependenciesFailed = true
					break
				}

				if taskResponses[dependency].Successful && taskResponses[dependency].Skipped {
					// skip this stage as the dependency was skipped ... don't run tasks that have dependencies that were skipped
					logger.Warn("Dependency skipped: '" + dependency + "' for stage: " + stage.Name + " skipping this stage")
					taskResponses[stage.Name] = data.TaskStatusResponse{TaskName: stage.Name, Successful: true, Skipped: true}
					break
				}
			}

			if dependenciesFailed {
				skippedTask := data.TaskStatusResponse{TaskName: stage.Name, Successful: false, Skipped: true}
				updatePipelineRun(skippedTask)
				continue
			}

			// if skipped because dependency skipped (response marked as skipped and dependencies haven't failed ^)
			if taskResponses[stage.Name].Skipped {
				skippedTask := data.TaskStatusResponse{TaskName: stage.Name, Successful: true, Skipped: true}
				updatePipelineRun(skippedTask)
				continue
			}
		}

		// if should skip this stage, break now and signal  ... if skipped by config mark as successful
		if stage.Skip {
			logger.Info("Skipping " + stage.Name + " based on config")
			taskResponses[stage.Name] = data.TaskStatusResponse{TaskName: stage.Name, Successful: true, Skipped: true}
			updatePipelineRun(taskResponses[stage.Name])
			continue
		}

		// run task
		go func(s data.Stage) {
			defer taskWg.Done()
			var start = time.Now()

			// Create a "running" status and update pipelineRun immediately
			runningTask := data.TaskStatusResponse{TaskName: s.Name, Successful: false, StartedAt: start}
			updatePipelineRun(runningTask)

			// spawn process to run task
			var successful, message = runTask(s, pipeline.Name)
			if !successful {
				logger.Error("Task failed: '" + s.Name + "' with message: " + message)
				taskStatusBuffer <- data.TaskStatusResponse{TaskName: s.Name, Successful: false, StartedAt: start, EndedAt: time.Now()}
			} else {
				taskStatusBuffer <- data.TaskStatusResponse{TaskName: s.Name, Successful: true, StartedAt: start, EndedAt: time.Now()}
			}
		}(stage)
		logger.Info("Running task: " + stage.Name)

		taskWg.Add(1)
		activeThreads++

		// if we have reached the maximum number of threads, wait for them to finish
		if activeThreads >= threads {
			logger.Debug("Pausing running new tasks, maximum threads reached")
			collect(true)
			logger.Debug("Resuming running new tasks")
		}
	}

	collect(false)

	pipelineRun.EndedAt = time.Now()
	for _, response := range taskResponses {
		updatePipelineRun(response) // this is probably redundant here, but safer to keep
		if !response.Successful {
			pipelineRun.Successful = false
		}
	}

	if !savePipelineRun(*pipelineRun, logger) {
		logger.Error("Error saving pipeline run for pipeline: " + pipeline.Name)
	}
	return pipelineRun.Successful, *pipelineRun
}
