package utils

import (
	"os"
)

type env string

func (e env) IsTest() bool {
	return os.Getenv("ENV") == "test"
}

func (e env) IsProd() bool {
	return os.Getenv("ENV") == "production"
}

func (e env) IsDev() bool {
	return os.Getenv("ENV") == "development"
}

func (e env) GetEnv() bool {
	return os.Getenv("ENV") == "development"
}

func Env() env {
	return "env"
}

