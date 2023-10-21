package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type env string

func (e *env) IsTest() bool {
	return os.Getenv("ENV") == "test"
}

func (e *env) IsProd() bool {
	return os.Getenv("ENV") == "production"
}

func (e *env) IsDev() bool {
	return os.Getenv("ENV") == "development"
}

func (e *env) GetEnv() bool {
	return os.Getenv("ENV") == "development"
}

func (e *env) SellTrue() bool {
	return len(os.Getenv("ALWAY_SELL")) > 0
}
func (e *env) SetModeTest() {
	os.Setenv("ENV", "test")
}

func Env() *env {
    e := env("env")
    return &e
}

func (e *env) getEnvNumber(envName string) float64 {
	number := os.Getenv(envName)
	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		panic(fmt.Errorf("trouble getting environment number %s %v", envName, err))
	}
	return value
}

func (e *env) QUOTE_BALANCE() float64 {
	return e.getEnvNumber("QUOTEST_BALANCE")
}

func (e *env) BASE_BALANCE() float64 {
	return e.getEnvNumber("BASETEST_BALANCE")
}

func (e *env) MaxInt() float64 {
	return e.getEnvNumber("MAX_INTEGER")
}
func (e *env) MinInt() float64 {
	return e.getEnvNumber("MIN_INTEGER")
}

func (e *env) RandomNumber() float64 {
	max := e.MaxInt() //, maxError := strconv.ParseFloat(os.Getenv("MAX_INTEGER"), 64)
	min := e.MinInt() //, minError := strconv.ParseFloat(os.Getenv("MIN_INTEGER"), 64)

	// if maxError != nil {
	// 	max = 100
	// }
	// if minError != nil {
	// 	min = 0
	// }

	if true {
		return RandomNumber(min, max)
	}
	v := rand.Intn(int(max)-int(min)) + int(min)
	return float64(v)
}

func RandomNumber(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
