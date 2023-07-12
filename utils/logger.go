package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"runtime"
)

func setupWrite(filename string) *os.File {
	var file *os.File

	if filename != "" {
		file, err := os.OpenFile( fmt.Sprintf("logs/%s%s", filename, ".log") , os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open log file")
		}
		fileWriter := zerolog.ConsoleWriter{Out: file}
		log.Logger = log.Output(fileWriter)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return file
}

func init() {
	
}

func LogInfo(s any) {
	pc, file, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	name := caller(file, line, details)
	defer setupWrite(name).Close()
	log.Info().Msg(s.(string))
}

func LogError(err error, s string) {
	log.Err(err).Msg(s)
}

func LogWarn(s string) {
	log.Warn().Msg(s)
}

func caller(file string, line int, details *runtime.Func) string {
	if details != nil {
		filename :=strings.ReplaceAll(filepath.Ext(details.Name()),".","")
		return fmt.Sprintf("%s_%d", filename, line)
	}
	return ""
}
