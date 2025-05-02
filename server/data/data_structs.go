package data

// TODO: should a stage support multiple tasks?
type Stage struct {
	Name      string   `json:"name"`
	Task      string   `json:"task"`
	DependsOn []string `json:"depends_on"`
	Pwd       string
	// TODO: should I add the ability to skip a given task
}

type Pipeline struct {
	Name         string
	Stages       []Stage
	Parallel     bool
	VariableFile string `json:"variable_file"`
}

type TaskStatusResponse struct {
	// TODO:
	// save start time, to track execution time
	// instead of bool success, use a status enum for returning to the UI
	TaskName   string
	Successful bool
}

type RegisteredPipeline struct {
	Name          string // the name of the pipeline, use as key
	Path          string // where the definition is stored
	VariablesFile string
	// TODO: should I add a list of runs here?
}
