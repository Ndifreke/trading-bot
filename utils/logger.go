package utils

import (
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LogInfo(s any) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	typeName := reflect.TypeOf(s).Name()
	switch typeName {
	case "string":
		log.Info().Msg(s.(string))
	case "":
	}
}

func LogError(err error, s string) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Err(err).Msg(s)
}


func LogWarn(s string) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Warn().Msg(s)
}

func GetCaller(){
	pc, a, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		fmt.Printf("called from %s\n %s",a, details.Name())
	}
}
