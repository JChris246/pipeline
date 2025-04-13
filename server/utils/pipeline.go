package utils

import (
	"encoding/json"
	"os"
	"path"
	"pipeline/data"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func validateVars(str string, variables map[string]string) []string {
	var missing []string

	re := regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)
	var results = re.FindAllStringSubmatch(str, -1)

	// no variables
	if results == nil {
		return missing
	}

	for _, result := range results {
		varName := result[1]
		if _, ok := variables[varName]; !ok {
			missing = append(missing, "Missing variable: "+varName)
		}
	}

	return missing
}

func ValidatePipelineDefinition(pipeline *data.Pipeline, vars *map[string]string, logger *logrus.Logger) []string {
	var errors []string

	// validate pipeline name
	if pipeline.Name == "" {
		logger.Error("Pipeline name is missing")
		errors = append(errors, "Pipeline name is missing")
	}

	// validate and load variable file
	var variables map[string]string
	if pipeline.VariableFile != "" && vars == nil {
		if _, err := os.Stat(pipeline.VariableFile); os.IsNotExist(err) {
			logger.Error("Variable file does not exist: " + pipeline.VariableFile)
			errors = append(errors, "Variable file does not exist: "+pipeline.VariableFile)
		} else {
			variables = LoadPipelineVars(pipeline.VariableFile, logger)
		}
	}

	if vars != nil {
		variables = *vars
	}

	// validate stages
	if len(pipeline.Stages) == 0 {
		logger.Error("Pipeline has no stages")
		errors = append(errors, "Pipeline has no stages")
	}

	var stageNames map[string]bool = make(map[string]bool) // make-shift set
	for i, stage := range pipeline.Stages {
		if stage.Name == "" {
			logger.Error("Stage name is missing at index " + strconv.Itoa(i))
			errors = append(errors, "Stage name is missing at stage index "+strconv.Itoa(i))
		}

		if stageNames[stage.Name] {
			logger.Error("Duplicate stage name: " + stage.Name + " at index " + strconv.Itoa(i))
			errors = append(errors, "Duplicate stage name: "+stage.Name+" at stage index "+strconv.Itoa(i))
		} else {
			stageNames[stage.Name] = true
		}

		// TODO: if support multiple tasks per stage update this check
		if stage.Task == "" {
			logger.Error(stage.Name + " ( index - " + strconv.Itoa(i) + ") stage task is missing")
			errors = append(errors, stage.Name+" ("+strconv.Itoa(i)+") stage task is missing")
		}

		var taskVariableErrors = validateVars(stage.Task, variables)
		if len(taskVariableErrors) > 0 {
			errors = append(errors, taskVariableErrors...)
		} else {
			pipeline.Stages[i].Task = injectVariables(stage.Task, variables)
		}

		var pwdVariableErrors = validateVars(stage.Task, variables)
		if len(pwdVariableErrors) > 0 {
			errors = append(errors, pwdVariableErrors...)
		} else {
			pipeline.Stages[i].Pwd = injectVariables(stage.Pwd, variables)
		}

		// since the intention is to run stages in the order they are defined, it should be fine to use the
		// stage name map in the current state to check for dependencies
		if len(stage.DependsOn) > 0 {
			for _, dependency := range stage.DependsOn {
				var _, dependencyExists = stageNames[dependency]
				if !dependencyExists {
					logger.Error(stage.Name + " ( index - " + strconv.Itoa(i) + ") depends on a non-existent stage: " + dependency)
					errors = append(errors, stage.Name+" ("+strconv.Itoa(i)+") dependency '"+dependency+"' has not been defined")
				}
			}
		}
	}

	return errors
}

func LoadDefinition(definitionPath string, logger *logrus.Logger) *data.Pipeline {
	if definitionPath == "" {
		logger.Error("Missing pipeline definition path")
		return nil
	}

	if _, err := os.Stat(definitionPath); os.IsNotExist(err) {
		logger.Error("Pipeline definition file does not exist: " + definitionPath)
		return nil
	}

	fileData, err := os.ReadFile(definitionPath)
	if err != nil {
		logger.Error("Error reading pipeline definition file: " + definitionPath)
		return nil
	}

	var pipeline data.Pipeline
	err = json.Unmarshal(fileData, &pipeline)
	if err != nil {
		logger.Error("Invalid pipeline definition file, unable to parse JSON: " + definitionPath)
		return nil
	}

	return &pipeline
}

func LoadPipelineVars(varFile string, logger *logrus.Logger) map[string]string {
	var variables = make(map[string]string)
	if varFile != "" {
		data, err := os.ReadFile(varFile)
		if err != nil {
			logger.Error("Error reading variable file: " + varFile)
			return nil
		}

		// TODO: support more formats for variable file, currently expects format like typical env files
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			// split line into key and value
			keyValuePair := strings.Split(line, "=")
			if len(keyValuePair) != 2 {
				logger.Warn("Invalid variable entry: " + line)
			} else {
				var key = strings.TrimSpace(keyValuePair[0])
				var value = strings.TrimSpace(keyValuePair[1])

				if key == "" {
					logger.Warn("Invalid variable entry - key is empty: " + line)
					continue
				}

				if value == "" {
					logger.Warn("Variable value is empty: " + line)
				}

				variables[key] = value // this will overwrite existing values
			}
		}
	}

	return variables
}

func injectVariables(task string, variables map[string]string) string {
	re := regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)
	return re.ReplaceAllStringFunc(task, func(match string) string {
		varName := match[1 : len(match)-1]
		return variables[varName]
	})
}

func CreateVariableFile(variables map[string]string, logger *logrus.Logger) string {
	InitDataStoreDir(logger)

	var fileId = GenerateId()
	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), "pipeline-variables-"+fileId+".properties")
	return SaveVariableFile(variables, filename, logger)
}

func SaveVariableFile(variables map[string]string, filepath string, logger *logrus.Logger) string {
	InitDataStoreDir(logger)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		logger.Error("Error creating variable file: " + err.Error())
		return ""
	}

	for key, value := range variables {
		_, err = file.WriteString(key + "=" + value + "\n")
		if err != nil {
			logger.Error("Error writing to variable file: " + err.Error())
			return ""
		}
	}

	err = file.Close()
	if err != nil {
		logger.Error("Error closing variable file: " + err.Error())
		return ""
	}

	return filepath
}

func SavePipelineDefinition(pipeline *data.Pipeline, logger *logrus.Logger) string {
	InitDataStoreDir(logger)

	var filename = path.Join(os.Getenv("DATA_STORE_DIR"), pipeline.Name+".json")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		logger.Error("Error creating pipeline definition file: " + err.Error())
		return ""
	}

	err = json.NewEncoder(file).Encode(pipeline)
	if err != nil {
		logger.Error("Error writing to pipeline definition file: " + err.Error())
		return ""
	}

	err = file.Close()
	if err != nil {
		logger.Error("Error closing pipeline definition file: " + err.Error())
		return ""
	}

	return filename
}
