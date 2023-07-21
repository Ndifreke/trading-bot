package utils

import (
	"math/rand"
	"os"
	"strconv"
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

func (e env) SellTrue() bool {
	return len(os.Getenv("ALWAY_SELL")) > 0
}

func Env() env {
	return "env"
}

func (e env) RandomNumber() float64 {
	max, maxError := strconv.ParseFloat(os.Getenv("MAX_INTEGER"), 64)
	min, minError := strconv.ParseFloat(os.Getenv("MIN_INTEGER"), 64)

	if maxError != nil {
		max = 100
	}
	if minError != nil {
		min = 0
	}
	v := rand.Intn(int(max)-int(min)) + int(min)
	return float64(v)
}
