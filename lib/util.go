package lib

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	value := GetEnv(key, strconv.Itoa(fallback))
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Println("Failed to read env var into an int")
		panic(err)
	}
	return intValue
}

func GetEnvList(key string, delimiter string) []string {
	value := GetEnv(key, "")
	if len(value) == 0 {
		return []string{}
	}
	return strings.Split(value, delimiter)
}
