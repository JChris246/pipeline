package data

import "time"

// TODO: should a stage support multiple tasks?
type Stage struct {
	Name      string   `json:"name"`
	Task      string   `json:"task"`
	DependsOn []string `json:"depends_on"`
	Pwd       string
	// TODO: should I add the ability to skip a given task?
}

type Pipeline struct {
	Name         string
	Stages       []Stage
	Parallel     bool
	VariableFile string `json:"variable_file"`
}

// TODO: do I need to convert these time.Time to int to save?
type TaskStatusResponse struct {
	// TODO: instead of bool success, use a status enum for returning to the UI?
	TaskName   string
	Successful bool
	Skipped    bool
	StartedAt  time.Time
	EndedAt    time.Time
}

type PipelineRun struct {
	Name       string
	Stages     []TaskStatusResponse
	StartedAt  time.Time
	EndedAt    time.Time
	Successful bool
	// TODO: should this store the logs for each task?
}

type RegisteredPipeline struct {
	Name          string // the name of the pipeline, use as key
	Path          string // where the definition is stored
	VariablesFile string
}

type PipelineItem struct {
	Name string
	// Stages   []Stage
	// Parallel bool
	Status  string
	LastRun int64
	Runs    []PipelineRun // store the last few runs
}
