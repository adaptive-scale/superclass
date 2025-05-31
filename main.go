package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Configure Logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// Set log level from environment variable or default to info
	logLevel := getEnvWithDefault("LOG_LEVEL", "info")
	if level, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Debug("Logrus initialized with JSON formatter")
}

func main() {
	log.Debug("Starting application initialization")

	// Log all environment variables in debug mode
	for _, env := range os.Environ() {
		log.WithField("env", env).Debug("Environment variable")
	}

	log.Info("Starting Superclass server")

	// Get configuration from environment
	log.Debug("Creating server instance from environment")
	server := NewServerFromEnv()

	// Get port from environment or use default
	port := getEnvIntWithDefault("PORT", 8080)

	log.WithFields(log.Fields{
		"port":        port,
		"pid":         os.Getpid(),
		"working_dir": getEnvWithDefault("PWD", "unknown"),
	}).Info("Server configuration loaded")

	log.Debug("Initiating server start sequence")
	// Start server
	if err := server.Start(port); err != nil {
		log.WithError(err).Fatal("Server failed to start")
	}
}
