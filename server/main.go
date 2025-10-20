package main

import (
	"flag"
	"fmt"
	"os"
	"pipeline/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const VERSION = "1.0.0"

func loadEnvVars(logger *logrus.Logger) bool {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
	}

	// use to stop execution if any required vars are missing; not currently used
	var missing bool = false

	// this would already have been loaded
	if os.Getenv("LOG_DIR") == "" {
		logger.Warn("Missing LOG_DIR environment variable")
		os.Setenv("LOG_DIR", "logs")
	}

	if os.Getenv("SERVER_PORT") == "" {
		logger.Warn("Missing SERVER_PORT environment variable")
	}

	if os.Getenv("ENV") == "" {
		logger.Warn("Missing ENV environment variable")
	}

	return missing
}

func run(logger *logrus.Logger, args []string) {
	logger.Info("Running headless")

	runCmd := flag.NewFlagSet("run", flag.ContinueOnError)
	definitionPath := runCmd.String("definition", "", "path to pipeline definition")
	// varFile := runCmd.String("variables", "", "path to variables file")
	runCmd.Parse(args[2:])

	// get definition file
	var pipeline = utils.LoadDefinition(*definitionPath, logger)
	if pipeline == nil {
		return
	}

	var errors = utils.ValidatePipelineDefinition(pipeline, nil, logger)

	if len(errors) > 0 {
		logger.Error("Pipeline validation failed")
		return
	}

	if success, _ := runPipeline(pipeline, nil, logger); success {
		logger.Info("Pipeline completed successfully")
	} else {
		logger.Error("Pipeline run failed")
	}
}

func help() {
	fmt.Println("Pipeline Tool " + VERSION)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  pipeline <subcommand> [options]")
	fmt.Println()
	fmt.Println("AVAILABLE SUBCOMMANDS:")
	fmt.Println("  run        Execute a pipeline definition file")
	fmt.Println("  serve      Start the pipeline server with web UI")
	fmt.Println("  version    Display the version information")
	fmt.Println("  help       Display this help message")
	fmt.Println()
	fmt.Println("RUN SUBCOMMAND OPTIONS:")
	fmt.Println("  -definition <path>    Path to pipeline definition file (required)")
	fmt.Println()
	fmt.Println("SERVE SUBCOMMAND:")
	fmt.Println("  Starts a web server for managing pipelines through a UI")
	fmt.Println("  Default port: 8080 (override with SERVER_PORT environment variable)")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  LOG_DIR       Directory for log files")
	fmt.Println("  SERVER_PORT   Port for the web server (default: 8080)")
	fmt.Println("  ENV          Environment mode")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  pipeline run --definition my-pipeline.json")
	fmt.Println("  pipeline serve")
	fmt.Println("  pipeline version")
	fmt.Println("  pipeline help")
}

func serve(logger *logrus.Logger) {
	logger.Info("Running as server")

	// Create a new Gin router
	router := gin.Default()
	gin.DefaultWriter = logger.Out
	if os.Getenv("ENV") == "prod" || os.Getenv("ENV") == "production" {
		// Set Gin mode to production
		// gin.SetMode(gin.ReleaseMode)

		// Disable console color
		// gin.DisableConsoleColor()

		// Enable recovery middleware
		router.Use(gin.Recovery())
	}

	initServer(logger)
	defineRoutes(router, logger)

	router.StaticFile("/", "static/index.html")
	router.Static("/assets", "static/assets")

	// Run the server
	var port = "8080"
	if os.Getenv("SERVER_PORT") != "" {
		port = os.Getenv("SERVER_PORT")
	}

	router.Run(":" + port)
}

func main() {
	// I know this order makes no sense
	logFile, logger := utils.SetupLogger("combined.log")

	if logFile != nil {
		defer logFile.Close()
	}

	if loadEnvVars(logger) {
		logger.Error("One or more dependant environment variables are missing.")
		return
	}

	if len(os.Args) < 2 {
		logger.Error("Missing subcommand")
		return
	}

	switch os.Args[1] {
	case "run":
		run(logger, os.Args)
	case "serve":
		serve(logger)
	case "version":
		logger.Info("Version: " + VERSION)
	case "help":
		help()
	default:
		logger.Error("Unknown subcommand: " + os.Args[1])
	}
}
