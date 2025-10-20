package utils

import (
	"os"
	"pipeline/data"
	"testing"
)

var _ = os.Setenv("ENV", "test")
var _, testLogger = SetupLogger("test.log")

// TODO: write more structured tests, with test cases and better error messages etc
// TODO: add more positive tests?

func Test_ValidatePipelineDefinition_ReturnsErrorForMissingPipelineName(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "echo"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertMin(t, 1, len(errors))
	AssertContains(t, errors, "Pipeline name is missing")
}

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

func Test_ValidatePipelineDefinition_ReturnsErrorWhenPassedVariablesAreNotSufficient_MultipleStageFields(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{
		Name: "stage 1",
		Task: "node {script}",
		Pwd:  "{workdir}",
		Args: []string{"{arg1}", "{arg2}"},
		Env:  []string{"PATH={env_path}"},
	})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &map[string]string{}, testLogger)

	// assert
	AssertMin(t, 5, len(errors))
	AssertContains(t, errors, "Missing variable: script")
	AssertContains(t, errors, "Missing variable: workdir")
	AssertContains(t, errors, "Missing variable: arg1")
	AssertContains(t, errors, "Missing variable: arg2")
	AssertContains(t, errors, "Missing variable: env_path")
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

func Test_ValidatePipelineDefinition_InjectsVariablesInAllStageFields(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{
		Name: "stage 1",
		Task: "node {script}",
		Pwd:  "{workdir}",
		Args: []string{"{arg1}", "{arg2}"},
		Env:  []string{"PATH={env_path}"}})

	var variables = map[string]string{
		"script":   "app.js",
		"workdir":  "/home/user",
		"arg1":     "start",
		"arg2":     "production",
		"env_path": "/usr/bin"}

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &variables, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "node app.js", pipeline.Stages[0].Task)
	AssertStringEqual(t, "/home/user", pipeline.Stages[0].Pwd)
	AssertStringEqual(t, "start", pipeline.Stages[0].Args[0])
	AssertStringEqual(t, "production", pipeline.Stages[0].Args[1])
	AssertStringEqual(t, "PATH=/usr/bin", pipeline.Stages[0].Env[0])
}

func Test_ValidatePipelineDefinition_HandlesCaseSensitiveVariables(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "echo {VAR} {var} {Var}"})
	var variables = map[string]string{"VAR": "uppercase", "var": "lowercase", "Var": "mixed"}

	// act
	var errors = ValidatePipelineDefinition(&pipeline, &variables, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
	AssertStringEqual(t, "echo uppercase lowercase mixed", pipeline.Stages[0].Task)
}

func Test_ValidatePipelineDefinition_ReturnsNoErrorsForValidPipelineWithMinimalData(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test pipeline", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage 1", Task: "echo 'hello'"})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_ValidatePipelineDefinition_HandlesStagesWithEmptyArgsAndEnv(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "simple", Task: "echo hello", Args: []string{}, Env: []string{}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_ValidatePipelineDefinition_HandlesMultipleDependenciesCorrectly(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "build", Task: "npm run build"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "test", Task: "npm test"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "deploy", Task: "npm run deploy", DependsOn: []string{"build", "test"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_ValidatePipelineDefinition_ReturnsErrorForPartialDependencyMatches(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "build", Task: "npm run build"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "deploy", Task: "npm run deploy", DependsOn: []string{"build", "test", "lint"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 2, len(errors))
	AssertContains(t, errors, "deploy (1) dependency 'test' has not been defined")
	AssertContains(t, errors, "deploy (1) dependency 'lint' has not been defined")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForSelfDependency(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "stage1", Task: "echo", DependsOn: []string{"stage1"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "stage1 (0) listed self as dependency")
}

func Test_ValidatePipelineDefinition_ReturnsErrorForMultipleDependenciesIncludingSelf(t *testing.T) {
	// arrange
	var pipeline = data.Pipeline{Name: "test", Stages: []data.Stage{}}
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "build", Task: "npm run build"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "test", Task: "npm test"})
	pipeline.Stages = append(pipeline.Stages, data.Stage{Name: "deploy", Task: "npm run deploy", DependsOn: []string{"build", "test", "deploy"}})

	// act
	var errors = ValidatePipelineDefinition(&pipeline, nil, testLogger)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "deploy (2) listed self as dependency")
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

func Test_validateVars_HandlesDuplicateVariables(t *testing.T) {
	// arrange
	var variables = map[string]string{"root": "/home"}
	var str = "cp {root}/file1 {root}/file2 {root}/backup"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_HandlesVariablesWithNumbers(t *testing.T) {
	// arrange
	var variables = map[string]string{"var1": "value1", "var2": "value2"}
	var str = "command {var1} {var2}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_HandlesVariablesWithUnderscores(t *testing.T) {
	// arrange
	var variables = map[string]string{"my_var": "value", "another_var_123": "value2"}
	var str = "command {my_var} {another_var_123}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_IgnoresMalformedVariables(t *testing.T) {
	// arrange
	var variables = map[string]string{"valid": "value"}
	var str = "command {valid} {invalid-var} {another.var} {space var}"

	// act
	var errors = validateVars(str, variables)

	// assert
	// Should only report missing "valid" variables, malformed ones are ignored by regex
	AssertEqual(t, 0, len(errors))
}

func Test_validateVars_HandlesVariableAtStringStart(t *testing.T) {
	// arrange
	var variables = map[string]string{}
	var str = "{start_var}/path/to/file"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: start_var")
}

func Test_validateVars_HandlesVariableAtStringEnd(t *testing.T) {
	// arrange
	var variables = map[string]string{}
	var str = "/path/to/{end_var}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: end_var")
}

func Test_validateVars_HandlesConsecutiveVariables(t *testing.T) {
	// arrange
	var variables = map[string]string{"var1": "value1"}
	var str = "{var1}{missing_var}"

	// act
	var errors = validateVars(str, variables)

	// assert
	AssertEqual(t, 1, len(errors))
	AssertContains(t, errors, "Missing variable: missing_var")
}

func Test_validateVars_HandlesNilVariablesMap(t *testing.T) {
	// arrange
	var str = "command {var1} {var2}"

	// act
	var errors = validateVars(str, nil)

	// assert
	// Should handle nil map gracefully and report all variables as missing
	AssertEqual(t, 2, len(errors))
	AssertContains(t, errors, "Missing variable: var1")
	AssertContains(t, errors, "Missing variable: var2")
}

// TODO: update the logic to only add unique missing vars to the error list?
func Test_validateVars_ReportsEachMissingVariableOnce(t *testing.T) {
	// arrange
	var variables = map[string]string{}
	var str = "command {missing} and {missing} again {missing}"

	// act
	var errors = validateVars(str, variables)

	// assert
	// Should report the same missing variable multiple times if it appears multiple times
	AssertEqual(t, 3, len(errors))
	AssertContains(t, errors, "Missing variable: missing")
}

func Test_validateVars_HandlesEmptyVariableName(t *testing.T) {
	// arrange
	var variables = map[string]string{"valid": "value"}
	var str = "command {valid} {}"

	// act
	var errors = validateVars(str, variables)

	// assert
	// Empty braces {} should be ignored by the regex pattern
	AssertEqual(t, 0, len(errors))
}

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

func Test_loadPipelineVars_ShouldReturnEmptyMapWhenFilePathIsEmpty(t *testing.T) {
	// arrange
	var filePath = ""

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 0, len(variables))
}

func Test_loadPipelineVars_ShouldHandleInvalidFormatLines_OnlyLoadValidVars(t *testing.T) {
	// arrange
	var filePath = "../test_assets/test_invalid_var_file.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "value1", variables["key1"])
	AssertStringEqual(t, "value2", variables["key2"])
}

func Test_loadPipelineVars_ShouldHandleEmptyKeys(t *testing.T) {
	// arrange
	var filePath = "../test_assets/missing_key_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 1, len(variables))
	AssertStringEqual(t, "value2", variables["valid_key"])
}

func Test_loadPipelineVars_ShouldHandleEmptyValues(t *testing.T) {
	// arrange
	var filePath = "../test_assets/empty_var_value.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "", variables["key1"])
	AssertStringEqual(t, "value2", variables["key2"])
}

func Test_loadPipelineVars_ShouldTrimWhitespace(t *testing.T) {
	// arrange
	var filePath = "../test_assets/trailing_space_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "value1", variables["key1"])
	AssertStringEqual(t, "value2", variables["key2"])
}

func Test_loadPipelineVars_ShouldOverwriteDuplicateKeys(t *testing.T) {
	// arrange
	var filePath = "../test_assets/duplicate_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "second_value", variables["key1"])
	AssertStringEqual(t, "value2", variables["key2"])
}

func Test_loadPipelineVars_ShouldHandleMultipleEquals(t *testing.T) {
	// arrange
	var filePath = "../test_assets/multiple_equals_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "value1", variables["key1"])
	AssertStringEqual(t, "value3", variables["key3"])
}

func Test_loadPipelineVars_ShouldHandleEmptyLines(t *testing.T) {
	// arrange
	var filePath = "../test_assets/empty_lines_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 2, len(variables))
	AssertStringEqual(t, "value1", variables["key1"])
	AssertStringEqual(t, "value2", variables["key2"])
}

func Test_loadPipelineVars_ShouldHandleSpecialCharacters(t *testing.T) {
	// arrange
	var filePath = "../test_assets/special_chars_var.txt"

	// act
	var variables = LoadPipelineVars(filePath, testLogger)

	// assert
	AssertEqual(t, 3, len(variables))
	AssertStringEqual(t, "/usr/bin:/usr/local/bin", variables["PATH"])
	AssertStringEqual(t, "https://example.com/path", variables["SIMPLE_URL"])
	AssertStringEqual(t, "echo \"hello world\"", variables["COMMAND"])
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

func Test_validateKeyValuePair_ValidFormat(t *testing.T) {
	// arrange
	validLine := "KEY=value"

	// act
	result, error := validateKeyValuePair(validLine)

	// assert
	if error != "" {
		t.Errorf("validateKeyValuePair() should not return error for valid format: %s", error)
	}

	if result != "KEY=value" {
		t.Errorf("validateKeyValuePair() should return trimmed line, got: %s", result)
	}
}

func Test_validateKeyValuePair_ValidFormatWithSpaces(t *testing.T) {
	// arrange
	validLineWithSpaces := "  KEY=value  "

	// act
	result, error := validateKeyValuePair(validLineWithSpaces)

	// assert
	if error != "" {
		t.Errorf("validateKeyValuePair() should not return error for valid format with spaces: %s", error)
	}

	if result != "KEY=value" {
		t.Errorf("validateKeyValuePair() should return trimmed line, got: %s", result)
	}
}

func Test_validateKeyValuePair_EmptyValue(t *testing.T) {
	// arrange
	emptyValueLine := "KEY="

	// act
	result, error := validateKeyValuePair(emptyValueLine)

	// assert
	if error != "" {
		t.Errorf("validateKeyValuePair() should not return error for empty value: %s", error)
	}

	if result != "KEY=" {
		t.Errorf("validateKeyValuePair() should return trimmed line, got: %s", result)
	}
}

func Test_validateKeyValuePair_InvalidFormat_NoEquals(t *testing.T) {
	// arrange
	invalidLine := "KEYVALUE"

	// act
	result, error := validateKeyValuePair(invalidLine)

	// assert
	if error == "" {
		t.Error("validateKeyValuePair() should return error for missing equals sign")
	}

	expectedError := "invalid env format: 'KEYVALUE'"
	if error != expectedError {
		t.Errorf("validateKeyValuePair() should return correct error message, got: %s", error)
	}

	if result != "" {
		t.Errorf("validateKeyValuePair() should return empty string for invalid format, got: %s", result)
	}
}

func Test_validateKeyValuePair_InvalidFormat_WrongSeparator(t *testing.T) {
	// arrange
	invalidLine := "KEY:VALUE"

	// act
	result, error := validateKeyValuePair(invalidLine)

	// assert
	if error == "" {
		t.Error("validateKeyValuePair() should return error for missing equals sign")
	}

	expectedError := "invalid env format: 'KEY:VALUE'"
	if error != expectedError {
		t.Errorf("validateKeyValuePair() should return correct error message, got: %s", error)
	}

	if result != "" {
		t.Errorf("validateKeyValuePair() should return empty string for invalid format, got: %s", result)
	}
}

func Test_validateKeyValuePair_InvalidFormat_MultipleEquals(t *testing.T) {
	// arrange
	invalidLine := "KEY=VALUE=EXTRA"

	// act
	result, error := validateKeyValuePair(invalidLine)

	// assert
	if error == "" {
		t.Error("validateKeyValuePair() should return error for multiple equals signs")
	}

	expectedError := "invalid env format: 'KEY=VALUE=EXTRA'"
	if error != expectedError {
		t.Errorf("validateKeyValuePair() should return correct error message, got: %s", error)
	}

	if result != "" {
		t.Errorf("validateKeyValuePair() should return empty string for invalid format, got: %s", result)
	}
}

func Test_validateKeyValuePair_InvalidFormat_EmptyString(t *testing.T) {
	// arrange
	emptyLine := ""

	// act
	result, error := validateKeyValuePair(emptyLine)

	// assert
	if error == "" {
		t.Error("validateKeyValuePair() should return error for empty string")
	}

	expectedError := "invalid env format: ''"
	if error != expectedError {
		t.Errorf("validateKeyValuePair() should return correct error message, got: %s", error)
	}

	if result != "" {
		t.Errorf("validateKeyValuePair() should return empty string for invalid format, got: %s", result)
	}
}

// function should maybe handle this as error, but we'll let cmd.Env handle it
func Test_validateKeyValuePair_InvalidFormat_OnlyEquals(t *testing.T) {
	// arrange
	equalsOnlyLine := "="

	// act
	result, error := validateKeyValuePair(equalsOnlyLine)

	// assert
	if error != "" {
		t.Errorf("validateKeyValuePair() should not return error for '=' (empty key and value): %s", error)
	}

	if result != "=" {
		t.Errorf("validateKeyValuePair() should return trimmed line, got: %s", result)
	}
}

func Test_validateKeyValuePair_ValidFormatWithSpecialCharacters(t *testing.T) {
	// arrange
	specialCharsLine := "PATH=/usr/bin:/usr/local/bin"

	// act
	result, error := validateKeyValuePair(specialCharsLine)

	// assert
	if error != "" {
		t.Errorf("validateKeyValuePair() should not return error for valid format with special chars: %s", error)
	}

	if result != "PATH=/usr/bin:/usr/local/bin" {
		t.Errorf("validateKeyValuePair() should return trimmed line, got: %s", result)
	}
}
