package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func logToFile(filename string) {
	var file *os.File
	file, err := os.OpenFile(fmt.Sprintf("logs/%s%s", filename, ".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}
	fileWriter := zerolog.ConsoleWriter{Out: file}
	log.Logger = log.Output(fileWriter)
}

func createLog(filename string) *os.File {
	var file *os.File
	isFileLogging := os.Getenv("FILE_LOGGING")
	if isFileLogging != "" {
		logToFile(filename)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return file
}

func caller(file string, line int, details *runtime.Func) string {
	if details != nil {
		filename := strings.ReplaceAll(filepath.Ext(details.Name()), ".", "")
		return fmt.Sprintf("%s_%d", filename, line)
	}
	return ""
}

func writeLog(e zerolog.Event, m string) {
	e.Msg(m)
}

func LogInfo(s any) {
	pc, file, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	name := caller(file, line, details)
	createLog(name).Close()
	log.Info().Msg(s.(string))
}

func LogError(err error, s string) {
	pc, file, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	name := caller(file, line, details)
	createLog(name).Close()
	log.Err(err).Msg(s)
}

func LogWarn(s string) {
	pc, file, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	name := caller(file, line, details)
	createLog(name).Close()
	log.Warn().Msg(s)
}

