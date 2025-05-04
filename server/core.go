package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"pipeline/data"
	"pipeline/utils"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func runTask(stage data.Stage, pipelineName string) (bool, string) {
	var args = strings.Split(stage.Task, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = stage.Pwd

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

	var errorLogName = utils.CreateOutputLogName(pipelineName, stage.Name, true)
	errorLogFile, err := os.OpenFile(errorLogName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err.Error()
	}
	defer errorLogFile.Close()

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

// TODO: maybe pass data.PipelineRun to be able to monitor
func runPipeline(pipeline *data.Pipeline, logger *logrus.Logger) (bool, data.PipelineRun) {
	var threads = 1
	if pipeline.Parallel {
		threads = runtime.NumCPU() / 2 // should this be configurable?
	}
	logger.Debug("Running pipeline " + pipeline.Name + " with " + fmt.Sprint(threads) + " thread(s)")

	var pipelineRun = data.PipelineRun{
		Name:      pipeline.Name,
		StartedAt: time.Now(),
	}

	// this will need to be managed better when the UI is added and pipelines can be registered
	taskResponses := make(map[string]data.TaskStatusResponse, len(pipeline.Stages))
	taskStatusBuffer := make(chan data.TaskStatusResponse, threads)

	var taskWg sync.WaitGroup
	var activeThreads = 0
	for _, stage := range pipeline.Stages {
		// making tasks wait on their dependencies. this why it is critical to define
		// the order of tasks carefully, because this will block all subsequent tasks
		if len(stage.DependsOn) > 0 {
			logger.Debug("Checking dependencies for: " + stage.Name)
			var dependenciesFailed = false
			for _, dependency := range stage.DependsOn {
				if _, ok := taskResponses[dependency]; !ok {
					logger.Info("Waiting for dependency: '" + dependency + "' for stage: " + stage.Name)
					taskWg.Wait() // wait for dependency to finish
					close(taskStatusBuffer)
					for i := range taskStatusBuffer {
						taskResponses[i.TaskName] = i
					}
					taskStatusBuffer = make(chan data.TaskStatusResponse, threads)
					activeThreads = 0
					logger.Info("Dependency finished: '" + dependency + "' for stage: " + stage.Name)
				}

				if !taskResponses[dependency].Successful {
					logger.Warn("Dependency failed: '" + dependency + "' for stage: " + stage.Name + " skipping this stage")
					// skip this stage as the dependency failed, also mark this stage as failed
					taskResponses[stage.Name] = data.TaskStatusResponse{TaskName: stage.Name, Successful: false, Skipped: true}
					dependenciesFailed = true
					break
				}
			}

			if dependenciesFailed {
				continue
			}
		}

		// run task
		go func(s data.Stage) {
			defer taskWg.Done()
			var start = time.Now()

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
			// maybe this should be extracted for reuse?
			taskWg.Wait()
			close(taskStatusBuffer)
			for i := range taskStatusBuffer {
				taskResponses[i.TaskName] = i
			}
			taskStatusBuffer = make(chan data.TaskStatusResponse, threads)
			activeThreads = 0
			logger.Debug("Resuming running new tasks")
		}
	}

	taskWg.Wait()
	close(taskStatusBuffer)
	for i := range taskStatusBuffer {
		taskResponses[i.TaskName] = i
	}

	pipelineRun.EndedAt = time.Now()
	pipelineRun.Successful = true
	for _, response := range taskResponses {
		pipelineRun.Stages = append(pipelineRun.Stages, response)
		if !response.Successful {
			pipelineRun.Successful = false
		}
	}

	if !savePipelineRun(pipelineRun, logger) {
		logger.Error("Error saving pipeline run for pipeline: " + pipeline.Name)
	}
	return pipelineRun.Successful, pipelineRun
}
