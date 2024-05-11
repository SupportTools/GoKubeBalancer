package logging

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
)

var logger *logrus.Logger

// SetupLogging initializes the logger with specific formatting and level.
func SetupLogging() *logrus.Logger {
	logger = logrus.New()
	logger.SetReportCaller(true)

	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		filename := getRelativePath(f.File)
		return "", filename + ":" + strconv.Itoa(f.Line)
	}
	logger.SetFormatter(customFormatter)

	logger.SetOutput(os.Stderr)

	if config.CFG.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

// GetRelativePath returns the file path relative to the project's root directory.
func getRelativePath(filePath string) string {
	wd, err := os.Getwd()
	if err != nil {
		logger.Errorf("Unable to get current working directory: %v", err)
		return filePath
	}
	relPath, err := filepath.Rel(wd, filePath)
	if err != nil {
		logger.Errorf("Unable to get relative path for %s: %v", filePath, err)
		return filePath
	}
	return relPath
}
