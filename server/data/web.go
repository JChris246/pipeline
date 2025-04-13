package data

type ApiErrorResponse struct {
	Message string `json:"msg"`
}

type RegisterFilePath struct {
	DefinitionFilePath string `json:"filepath"`
	VariableFilePath   string `json:"variable_file"`
}

type RegisterPipelineRequest struct {
	PipelineDefinition Pipeline          `json:"pipeline"`
	Variables          map[string]string `json:"variables"`
}

type RegisteredPipelineResponse struct {
	Name    string `json:"name"`     // the name of the pipeline, use as key
	LastRun int64  `json:"last_run"` // the last time the pipeline was run
	Status  string `json:"status"`   // the current status of the pipeline
}

type RegisteredPipelineDetails struct {
	Name      string            `json:"name"`
	Stages    []Stage           `json:"stages"`
	Parallel  bool              `json:"parallel"`
	Variables map[string]string `json:"variables"`
	LastRun   int64             `json:"last_run"` // the last time the pipeline was run
	Status    string            `json:"status"`   // the current status of the pipeline
	// TODO: should I add a list of run here?
	// TODO: add last run logs
}

type EditPipelineRequest struct {
	Name      string            `json:"name"`
	Stages    []Stage           `json:"stages"`
	Parallel  bool              `json:"parallel"`
	Variables map[string]string `json:"variables"`
}
