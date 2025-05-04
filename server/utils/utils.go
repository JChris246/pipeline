package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x%x-%d", b[0:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:], time.Now().UnixMilli())
}

// GetCurrentTimeStamp generates a timestamp in the format "YYYY-MM-DD HH_MM" or "YYYY-MM-DD HH_MM_SS" if seconds is true.
//
// seconds: If true, includes the seconds in the timestamp.
// string: The generated timestamp.
func GetCurrentTimeStamp(seconds bool) string {
	currentDate := time.Now()
	date := fmt.Sprintf("%d-%02d-%02d", currentDate.Year(), currentDate.Month(), currentDate.Day())
	time := fmt.Sprintf("%02d_%02d", currentDate.Hour(), currentDate.Minute())
	if seconds {
		time += fmt.Sprintf("_%02d", currentDate.Second())
	}
	return strings.Join([]string{date, time}, " ")
}

// CreateOutputLogName generates the name for an output log file.
//
// Parameters:
// - pipelineName: a string representing the name of the pipeline which will be the parent directory of all the pipeline log files.
// - stageName: a string representing the name of the stage which will be part of the log file name.
// - error: a boolean indicating whether the log is for an error or not.
//
// Return type:
// - string: the generated output log file name.
func CreateOutputLogName(pipelineName string, stageName string, error bool) string {
	timestamp := GetCurrentTimeStamp(true)
	logDir := filepath.Join(os.Getenv("LOG_DIR"), pipelineName)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, os.ModePerm)
	}
	fileName := timestamp + " " + stageName + "-" + map[bool]string{true: "stderr", false: "stdout"}[error] + ".txt"
	return filepath.Join(os.Getenv("LOG_DIR"), pipelineName, fileName)
}

// getNextArchiveNumber returns the next available archive number as a string.
//
// It does this by checking the existence of files in the log directory with a specific naming pattern.
// The function starts with an initial archive number of 1 and increments it until it finds a number
// that does not correspond to an existing file in the log directory.
//
// Return:
// - The next available archive number as a string.
func getNextArchiveNumber(logName string) string {
	current := 1
	logDir := filepath.Join(os.Getenv("LOG_DIR"), logName+".")

	for {
		filePath := logDir + strconv.Itoa(current)
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			break
		}
		current++
	}

	return strconv.Itoa(current)
}

// SetupLogger sets up the logger based on the environment.
//
// It loads the environment from a .env file, sets the log level based on the environment,
// and configures the logger to output to either stdout or a file depending on the environment.
//
// Returns a file pointer and a logrus logger instance.
func SetupLogger(logName string) (*os.File, *logrus.Logger) {
	// Load env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	var env = os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	if env == "test" {
		// do we actually want to log, if in test mode?
		var testLogger = logrus.New()
		testLogger.SetLevel(logrus.InfoLevel)
		testLogger.SetOutput(os.Stdout)
		testLogger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
		return nil, testLogger
	}

	// create dir for the log file if it doesn't yet exist
	if _, err := os.Stat(os.Getenv("LOG_DIR")); os.IsNotExist(err) {
		err2 := os.MkdirAll(os.Getenv("LOG_DIR"), os.ModePerm)
		if err2 != nil {
			fmt.Println("Error creating log directory: " + os.Getenv("LOG_DIR"))
			fmt.Println(err)
		} else {
			fmt.Println("Created log directory")
		}
	}

	// setup logger
	var Logger = logrus.New()
	var isProd = env == "prod" || env == "production"
	if isProd {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		// Logger.SetLevel(logrus.InfoLevel)
		Logger.SetLevel(logrus.DebugLevel)
	}
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true, // !isProd,
		FullTimestamp: true,
	})

	// archive current log files if exist
	logFilePath := filepath.Join(os.Getenv("LOG_DIR"), logName)
	_, statErr := os.Stat(logFilePath)
	if statErr == nil {
		fmt.Println("Found existing log file: " + logFilePath)
		var archiveNumber = getNextArchiveNumber(logName)
		renameError := os.Rename(logFilePath, logFilePath+"."+archiveNumber)
		if renameError != nil {
			fmt.Println("Error renaming existing log file: " + renameError.Error())
		}
	}

	// create log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)

	// setup outputs for the logger
	if isProd && err == nil {
		Logger.SetOutput(file)
		return file, Logger
	} else if isProd && err != nil {
		Logger.SetOutput(os.Stdout)
		Logger.Error("Failed to setup logger with file: " + err.Error())
		return nil, Logger
	} else if !isProd && err == nil {
		// Log to the file in addition to stdout
		Logger.SetOutput(io.MultiWriter(os.Stdout, file))
		return file, Logger
	} else {
		Logger.SetOutput(os.Stdout)
		Logger.Warn(err)
		Logger.Warn("Unable to setup logger with a file")
		return nil, Logger
	}
}

func InitDataStoreDir(logger *logrus.Logger) {
	if _, err := os.Stat(os.Getenv("DATA_STORE_DIR")); os.IsNotExist(err) {
		err2 := os.MkdirAll(os.Getenv("DATA_STORE_DIR"), os.ModePerm)
		if err2 != nil {
			logger.Error("Error creating log directory: " + os.Getenv("DATA_STORE_DIR") + " - " + err2.Error())
		} else {
			logger.Debug("Log directory created: " + os.Getenv("DATA_STORE_DIR"))
		}
	}
}

func InitDir(dir string, logger *logrus.Logger) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err2 := os.MkdirAll(dir, os.ModePerm)
		if err2 != nil {
			logger.Error("Error creating directory: " + dir + " - " + err2.Error())
			return false
		} else {
			logger.Debug("Directory created: " + dir)
			return true
		}
	}
	return true
}

func DeleteFile(filename string, logger *logrus.Logger) bool {
	if filename != "" {
		err := os.Remove(filename)
		if err != nil {
			logger.Error("Error deleting file '" + filename + "': " + err.Error())
			return false
		}
	}
	return true
}
