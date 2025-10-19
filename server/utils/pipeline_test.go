package utils

import (
	"os"
	"pipeline/data"
	"testing"
)

var _ = os.Setenv("ENV", "test")
var _, testLogger = SetupLogger("test.log")

// TODO: write more structured tests, with test cases and better error messages etc
// TODO: add positive tests

func Test_ValidatePipelineDefinition_ReturnsErrorForHavingZeroStages(t *testing.T) {
	// act
	var errors = ValidatePipelineDefinition(&data.Pipeline{Stages: []data.Stage{}}, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Pipeline has no stages")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForHavingAStageWithNoName(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "", Task: "echo"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	if len(errors) < 1 {
		t.Errorf("Expected error for having a stage with no name")
	}
	AssertContains(t, errors, "Stage name is missing at stage index 0")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForHavingAStageWithNoTask(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "stage 1 (0) stage task is missing")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForHavingAStageWithDuplicateName(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 2"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Duplicate stage name: stage 1 at stage index 2")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForHavingAStageWithNonExistentDependency(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "echo", DependsOn: []string{"stage 3"}})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 2", Task: "echo"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "stage 1 (0) dependency 'stage 3' has not been defined")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForNonExistentVariableFile(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}, VariableFile: "does/not/exist"}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "echo"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Variable file does not exist: does/not/exist")
}

// TODO: add tests for ValidatePipelineDefinition where vars are passed
func Test_ValidatePipelineDefinition_ReturnsErrorWhenPassedVariablesAreNotSufficient(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node {root}/media_central_index.js", Pwd: "{root}"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: root")
}

func Test_ValidatePipelineDefinition_ReturnsErrorWhenPassedVariablesAreNotSufficientForPwd(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node /home/media_central_index.js", Pwd: "{root}"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: root")
}

func Test_ValidatePipelineDefinition_ReturnsErrorWhenPassedVariablesAreNotSufficientForArgs(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Args: []string{"/home/media_central_index.js", "{action}"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: action")
}

func Test_ValidatePipelineDefinition_ReturnsErrorWhenPassedVariablesAreNotSufficientForEnv(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Env: []string{"FFMPEG_PATH={ffmpeg_path}"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: ffmpeg_path")
}

func Test_ValidatePipelineDefinition_ReturnsNoErrorsWhenPassedVariablesAreSufficient(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node {root}/media_central_index.js", Pwd: "{root}"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{"root": "/root"}, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "node /root/media_central_index.js", pipeline.Stages[0].Task)
	AssertStringEqual(t, "/root", pipeline.Stages[0].Pwd)
}

func Test_ValidatePipelineDefinition_ReturnsNoErrorsWhenPassedVariablesAreSufficientForPwd(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node /home/media_central_index.js", Pwd: "{root}"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{"root": "/root"}, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "/root", pipeline.Stages[0].Pwd)
}

func Test_ValidatePipelineDefinition_ReturnsNoErrorsWhenPassedVariablesAreSufficientForArgs(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Args: []string{"/home/media_central_index.js", "{action}"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{"action": "add"}, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "add", pipeline.Stages[0].Args[1])
}

func Test_ValidatePipelineDefinition_ReturnsNoErrorsWhenPassedVariablesAreSufficientForEnv(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Env: []string{"FFMPEG_PATH={ffmpeg_path}"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{"ffmpeg_path": "/usr/bin"}, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "FFMPEG_PATH=/usr/bin", pipeline.Stages[0].Env[0])
}

func Test_ValidatePipelineDefinition_ReturnsErrorWhenInvalidFormatForEnv(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Env: []string{"FFMPEG_PATH:/usr/bin"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "invalid env format: 'FFMPEG_PATH:/usr/bin'")
}

func Test_ValidatePipelineDefinition_ReturnsErrorWhenInvalidFormat_BlankString_ForEnv(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node", Env: []string{" "}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "invalid env format: ' '")
}

func Test_ValidatePipelineDefinition_ShouldUsePassedVariablesInsteadOfVariableFile(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "pipeline 1", Stages: []data.Stage{}, VariableFile: "../test_assets/test_var_file.txt"}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "node {root}/media_central_index.js", Pwd: "{root}"})
	var variables = map[string]string{"root": "/home/root/Downloads"}

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &variables, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "node /home/root/Downloads/media_central_index.js", pipeline.Stages[0].Task)
	AssertStringEqual(t, "/home/root/Downloads", pipeline.Stages[0].Pwd)
}

// testing pass by reference
func Test_ValidatePipelineDefinition_ReturnsPipelineWithInjectedVars(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "pipeline 1", Stages: []data.Stage{}, VariableFile: "../test_assets/test_var_file.txt"}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1",
		Task: "node {root}/media_central_index.js", Pwd: "{root}", Args: []string{"{root}"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "node /home/root/Documents/media_central_index.js", pipeline.Stages[0].Task)
	AssertStringEqual(t, "/home/root/Documents", pipeline.Stages[0].Pwd)
	AssertStringEqual(t, "/home/root/Documents", pipeline.Stages[0].Args[0])
}

func Test_validateVars_ReturnsErrorForNonExistentVariable(t *testing.T) {
	// arrange
	var variables = map[string]string{"varKey": "varValue"}
	var str = "node {root}/media_central_index.js.{suffix}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 2, len(errors))
	AssertContains(t, errors, "Missing variable: root")
	AssertContains(t, errors, "Missing variable: suffix")
}

func Test_validateVars_ReturnsErrorWhenSomeVarsDoNotExist(t *testing.T) {
	// arrange
	var variables = map[string]string{"suffix": "varValue"}
	var str = "node {root}/media_central_index.js.{suffix}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: root")
}

func Test_validateVars_ReturnsNoErrorsWhenAllVariablesExist(t *testing.T) {
	// arrange
	var variables = map[string]string{"root": "varValue", "suffix": "varValue"}
	var str = "node {root}/media_central_index.js.{suffix}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_ReturnsNoErrorsForEmptyString(t *testing.T) {
	var str = ""
	var variables = map[string]string{"varKey": "varValue"}
	var errors = validateVars(str, variables)
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_ReturnsNoErrorsForNoVariablesInString(t *testing.T) {
	var str = "hello world"
	var variables = map[string]string{"varKey": "varValue"}
	var errors = validateVars(str, variables)
	AssertEqual(t, 0, len(errors))
}

// TODO: could stand to add more tests for loadPipelineVars
func Test_loadPipelineVars_ShouldReturnEmptyMapWhenFileDoesNotExist(t *testing.T) {
	// arrange
	var filePath = "/does/not/exist"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 0, len(variables))
}

func Test_loadPipelineVars_ShouldCorrectlyLoadVarsFromFile(t *testing.T) {
	// arrange
	var filePath = "../test_assets/test_var_file.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertStringEqual(t, "/home/root/Documents", variables["root"])
	AssertStringEqual(t, "8086", variables["SERVER_PORT"])
	AssertStringEqual(t, "dev", variables["ENV"])
}

func Test_injectVariables_ShouldCorrectReplaceVariableKeyWithTheirValues(t *testing.T) {
	// arrange
	var task = "node {root}/media_central_index.js.{suffix}"
	var variables = map[string]string{"root": "/home/root/Documents", "suffix": "", "ENV": "dev"}

	// act
	var result = injectVariables(task, variables)

	// assert
	AssertStringEqual(t, "node /home/root/Documents/media_central_index.js.", result)
}
