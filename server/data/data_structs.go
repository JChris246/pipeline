package data

// TODO: should a stage support multiple tasks?
type Stage struct {
	Name      string
	Task      string
	DependsOn []string `json:"depends_on"`
	Pwd       string
}

type Pipeline struct {
	Name         string // this is not technically required unless you are using the UI and registering pipeline
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
